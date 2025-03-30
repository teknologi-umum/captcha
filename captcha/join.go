package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	"github.com/allegro/bigcache/v3"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// Captcha struct keeps all the data needed for the captcha
// for a certain user.
//
// It will be converted to JSON format (as array of bytes or []byte)
// and then will be stored to the in memory cache, with the key
// of the corresponding Telegram User ID.
type Captcha struct {
	// Store the correct answer for the captcha
	Answer string `json:"a"`
	// Expiry time for the captcha
	Expiry             time.Time `json:"e"`
	ChatID             int64     `json:"c"`
	SenderID           int64     `json:"s"`
	QuestionID         string    `json:"q"`
	AdditionalMessages []string  `json:"am"`
	UserMessages       []string  `json:"um"`
}

const (
	// BanDuration specifies how long a user will be banned in the group.
	BanDuration = 60 * time.Second
	// Timeout specifies how long the captcha question will be valid.
	// After this time, the user will be kicked.
	// Or banned exactly, for one hour.
	Timeout = 61 * time.Second
)

// DefaultQuestion contains the default captcha questions.
var DefaultQuestion = "Halo, {user}!\n\n" +
	"Sebelum lanjut, selesaikan captcha ini dulu agar bisa chat di grup ini. Ubah teks besar yang kamu lihat dibawah pesan ini menjadi teks biasa. Teks tersebut hanya berupa kombinasi angka 1-9 dengan huruf V, W, X, dan Y, jangan salah ketik ya!\n\n" +
	"Ini teksnya 👇, kamu punya waktu 1 menit dari sekarang! Kalau tulisannya pecah, dirotate layarnya kebentuk landscape ya.\n\n" +
	"<pre>{captcha}</pre>"

// CaptchaUserJoin is the most frustrating function that I've written
// at this point of time.
//
// As the function name says, it will prompt a captcha to the incoming user that
// has just joined the group.
//
// At the end of the function, it will create 2 goroutines in which
// both of them are responsible for kicking the user out of the group.
func (d *Dependencies) CaptchaUserJoin(ctx context.Context, m *tb.Message) {
	span := sentry.StartSpan(ctx, "captcha.user_join")
	defer span.Finish()
	ctx = span.Context()

	// Check if the user is an admin or bot first.
	// If they are, return.
	// If they're not, continue to execute the captcha.
	var admins []tb.ChatMember
	groupAdmins, err := d.Memory.Get("group-admins:" + strconv.FormatInt(m.Chat.ID, 10))
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			slog.DebugContext(ctx, "Setting cache entry for group admins", slog.Int64("group_id", m.Chat.ID))
			// Find and set
			admins, err = d.Bot.AdminsOf(ctx, m.Chat)
			if err != nil {
				if !strings.Contains(err.Error(), "Gateway Timeout (504)") && !strings.Contains(err.Error(), "retry after") {
					slog.ErrorContext(ctx, "failed to get group admins", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID))
					shared.HandleBotError(ctx, err, d.Bot, m)
					return
				}

				slog.WarnContext(ctx, "failed to get group admins", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID))
				shared.HandleBotError(ctx, err, d.Bot, m)
				return
			}

			var adminIDs []string
			for _, admin := range admins {
				adminIDs = append(adminIDs, strconv.FormatInt(admin.User.ID, 10))
			}

			groupAdmins = []byte(strings.Join(adminIDs, ","))

			err = d.Memory.Set("group-admins:"+strconv.FormatInt(m.Chat.ID, 10), groupAdmins)
			if err != nil {
				slog.ErrorContext(ctx, "failed to set group admins", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID))
				shared.HandleBotError(ctx, err, d.Bot, m)
				// DO NOT return, continue the captcha process.
			}
		} else {
			slog.ErrorContext(ctx, "failed to get group admins", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID))
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}
	} else {
		var adminIDs = bytes.Split(groupAdmins, []byte(","))
		for _, id := range adminIDs {
			parsedId, err := strconv.ParseInt(string(id), 10, 64)
			if err != nil {
				continue
			}
			admins = append(admins, tb.ChatMember{User: &tb.User{ID: parsedId}})
		}
	}

	if m.UserJoined.ID != 0 {
		m.Sender = m.UserJoined
	}

	if m.Sender.IsBot || m.Private() || utils.IsAdmin(admins, m.Sender) {
		slog.DebugContext(ctx, "User is a bot, private chat, or an admin, skipping captcha", slog.Int64("user_id", m.Sender.ID), slog.Int64("group_id", m.Chat.ID))
		return
	}

	// randNum generates a random number (3 digit) in string format
	var randNum = utils.GenerateRandomNumber()
	// captcha generates ascii art from the randNum value
	var captcha = utils.GenerateAscii(randNum)

	// Replacing the template from CaptchaQuestion
	question := strings.Replace(
		strings.Replace(DefaultQuestion, "{captcha}", captcha, 1),
		"{user}",
		"<a href=\"tg://user?id="+strconv.FormatInt(m.Sender.ID, 10)+"\">"+
			utils.SanitizeInput(m.Sender.FirstName)+utils.ShouldAddSpace(m.Sender)+utils.SanitizeInput(m.Sender.LastName)+
			"</a>",
		1,
	)

SENDMSG_RETRY:
	// Send the question first.
	msgQuestion, err := d.Bot.Send(
		ctx,
		m.Chat,
		question,
		&tb.SendOptions{
			ParseMode:             tb.ModeHTML,
			ReplyTo:               m,
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		},
	)
	if err != nil {
		var floodError tb.FloodError
		if errors.As(err, &floodError) {
			if floodError.RetryAfter == 0 {
				floodError.RetryAfter = 15
			}

			slog.DebugContext(ctx, "Received FloodError", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID), slog.Int64("user_id", m.Sender.ID), slog.Int("retry_after", floodError.RetryAfter))
			time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
			goto SENDMSG_RETRY
		}

		if strings.Contains(err.Error(), "Gateway Timeout (504)") {
			slog.DebugContext(ctx, "Received Gateway Timeout, retrying in 10 seconds", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID), slog.Int64("user_id", m.Sender.ID))
			time.Sleep(time.Second * 10)
			goto SENDMSG_RETRY
		}

		// err could possibly be nil at this point, so we better check it out.
		if err != nil {
			slog.ErrorContext(ctx, "Failed to send question", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID), slog.Int64("user_id", m.Sender.ID))
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}
	}

	// OK. We've sent the question. Now we are going to prepare the data that will
	// be kept on the in-memory store.
	//
	// The AdditionalMessages key will be added later when there is an additional message
	// sent by the bot.
	captchaData, err := json.Marshal(Captcha{
		Answer:             randNum,
		Expiry:             time.Now().Add(Timeout),
		ChatID:             m.Chat.ID,
		SenderID:           m.Sender.ID,
		QuestionID:         strconv.Itoa(msgQuestion.ID),
		AdditionalMessages: []string{strconv.Itoa(m.ID)},
		UserMessages:       nil,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal captcha data", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID), slog.Int64("user_id", m.Sender.ID))
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	// Yes, the cache key is their User ID in string format.
	err = d.DB.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(strconv.FormatInt(m.Chat.ID, 10)+":"+strconv.FormatInt(m.Sender.ID, 10)), captchaData)
		if err != nil {
			return err
		}

		var captchaUsers []byte

		captchaUsersItem, err := txn.Get([]byte("captcha:users:" + strconv.FormatInt(m.Chat.ID, 10)))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		if captchaUsersItem != nil && err == nil {
			captchaUsers, err = captchaUsersItem.ValueCopy(nil)
			if err != nil {
				return err
			}
		}

		err = txn.Set([]byte("captcha:users:"+strconv.FormatInt(m.Chat.ID, 10)), []byte(string(captchaUsers)+";"+strconv.FormatInt(m.Sender.ID, 10)))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save captcha data", slog.String("error", err.Error()), slog.Int64("group_id", m.Chat.ID), slog.Int64("user_id", m.Sender.ID))
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	d.waitOrDelete(ctx, m)
}
