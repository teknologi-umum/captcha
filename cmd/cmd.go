package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"teknologi-umum-captcha/analytics"
	"teknologi-umum-captcha/ascii"
	"teknologi-umum-captcha/badwords"
	"teknologi-umum-captcha/captcha"
	"teknologi-umum-captcha/shared"
	"teknologi-umum-captcha/underattack"
	"teknologi-umum-captcha/utils"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	tb "gopkg.in/telebot.v3"
)

// Dependency contains the dependency injection struct
// that is required for the main command to use.
//
// It will spread and use the correct dependencies for
// each packages on the captcha project.
type Dependency struct {
	Memory      *bigcache.BigCache
	Bot         *tb.Bot
	DB          *sqlx.DB
	Mongo       *mongo.Client
	MongoDBName string
	TeknumID    string
	captcha     *captcha.Dependencies
	ascii       *ascii.Dependencies
	analytics   *analytics.Dependency
	badwords    *badwords.Dependency
	underAttack *underattack.Dependency
}

// New returns a pointer struct of Dependency
// which map the incoming dependencies provided
// into what's needed by each domain.
func New(deps Dependency) *Dependency {
	analyticsDeps := &analytics.Dependency{
		Memory:   deps.Memory,
		Bot:      deps.Bot,
		DB:       deps.DB,
		TeknumID: deps.TeknumID,
	}
	return &Dependency{
		Bot:         deps.Bot,
		Memory:      deps.Memory,
		DB:          deps.DB,
		Mongo:       deps.Mongo,
		MongoDBName: deps.MongoDBName,
		TeknumID:    deps.TeknumID,
		captcha: &captcha.Dependencies{
			Memory:    deps.Memory,
			Bot:       deps.Bot,
			Analytics: analyticsDeps,
			TeknumID:  deps.TeknumID,
		},
		ascii: &ascii.Dependencies{
			Bot: deps.Bot,
		},
		analytics: analyticsDeps,
		badwords: &badwords.Dependency{
			Mongo:       deps.Mongo,
			MongoDBName: deps.MongoDBName,
		},
		underAttack: &underattack.Dependency{
			Memory: deps.Memory,
			DB:     deps.DB,
			Bot:    deps.Bot,
		},
	}
}

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.captcha.WaitForAnswer(ctx, c.Message())

	err := d.analytics.NewMessage(c.Message())
	if err != nil {
		shared.HandleError(ctx, err)
	}

	return nil
}

// OnUserJoinHandler handle any incoming user join,
// whether they were invited by someone (meaning they are
// added by someone else into the group), or they join
// the group all by themselves.
func (d *Dependency) OnUserJoinHandler(c tb.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	underAttack, err := d.underAttack.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleError(ctx, err)
	}

	if underAttack {
		err := d.underAttack.Kicker(c)
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, c.Message())
		}
		return nil
	}

	var tempSender *tb.User
	if c.Message().UserJoined.ID != 0 {
		tempSender = c.Message().UserJoined
	} else {
		tempSender = c.Message().Sender
	}

	go d.analytics.NewUser(ctx, c.Message(), tempSender)

	d.captcha.CaptchaUserJoin(ctx, c.Message())

	return nil
}

// OnNonTextHandler meant to handle anything else
// than an incoming text message.
func (d *Dependency) OnNonTextHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.captcha.NonTextListener(ctx, c.Message())

	err := d.analytics.NewMessage(c.Message())
	if err != nil {
		shared.HandleError(ctx, err)
	}

	return nil
}

// OnUserLeftHandler handles during an event in which
// a user left the group.
func (d *Dependency) OnUserLeftHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.captcha.CaptchaUserLeave(ctx, c.Message())
	return nil
}

// AsciiCmdHandler handle the /ascii command.
func (d *Dependency) AsciiCmdHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.ascii.Ascii(ctx, c.Message())
	return nil
}

// BadWordsCmdHandler handle the /badwords command.
// This can only be accessed by some users on Telegram
// and only valid for private chats.
func (d *Dependency) BadWordHandler(c tb.Context) error {
	if !c.Message().Private() {
		return nil
	}
	ok := d.badwords.Authenticate(strconv.FormatInt(c.Sender().ID, 10))
	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	err := d.badwords.AddBadWord(ctx, strings.TrimPrefix(c.Message().Text, "/badwords "))
	if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
		shared.HandleBotError(ctx, err, c.Bot(), c.Message())
		return nil
	}

	_, err = c.Bot().Send(c.Sender(), "Terimakasih telah menambahkan kata yang tidak pantas.")
	if err != nil {
		shared.HandleBotError(ctx, err, c.Bot(), c.Message())
	}

	return nil
}

// CukupHandler was created just to mock laode.
func (d *Dependency) CukupHandler(c tb.Context) error {
	if c.Message().Private() {
		return nil
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	_, err := c.Bot().Send(c.Chat(), &tb.Photo{File: tb.FromURL("https://i.ibb.co/WvynnPb/ezgif-4-13e23b17f1.jpg")})
	if err != nil {
		shared.HandleBotError(ctx, err, c.Bot(), c.Message())
	}

	return nil
}

// EnableUnderAttackModeHandler provides a handler for /underattack command.
func (d *Dependency) EnableUnderAttackModeHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.underAttack.EnableUnderAttackModeHandler(ctx, c)
}

// DisableUnderAttackModeHandler provides a handler for /disableunderattack command.
func (d *Dependency) DisableUnderAttackModeHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.underAttack.DisableUnderAttackModeHandler(ctx, c)
}

func (d *Dependency) SetirHandler(c tb.Context) error {
	admin := strings.Split(os.Getenv("ADMIN_ID"), ",")
	if !utils.IsIn(admin, strconv.FormatInt(c.Sender().ID, 10)) || c.Chat().Type != tb.ChatPrivate {
		return nil
	}

	home, err := strconv.ParseInt(d.TeknumID, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing teknum id: %w", err)
	}

	if c.Message().IsReply() {
		var replyToID int

		if strings.HasPrefix(c.Message().Payload, "https://t.me/") {
			replyToID, err = strconv.Atoi(strings.Split(c.Message().Payload, "/")[4])
			if err != nil {
				return err
			}
		} else {
			replyToID, err = strconv.Atoi(c.Message().Payload)
			if err != nil {
				return err
			}
		}

		_, err = d.Bot.Send(tb.ChatID(home), c.Message().ReplyTo.Text, &tb.SendOptions{
			ParseMode:         tb.ModeHTML,
			AllowWithoutReply: true,
			ReplyTo: &tb.Message{
				ID: replyToID,
				Chat: &tb.Chat{
					ID: int64(home),
				},
			},
		})
		if err != nil {
			_, err = d.Bot.Send(c.Chat(), "Failed sending that message: "+err.Error())
			if err != nil {
				return fmt.Errorf("failed sending that message: %w", err)
			}
		} else {
			_, err = d.Bot.Send(c.Chat(), "Message sent")
			if err != nil {
				return fmt.Errorf("sending message: %w", err)
			}
		}

		return nil
	}

	if strings.HasPrefix(c.Message().Payload, "https://") {
		var toBeSent interface{}
		if strings.HasSuffix(c.Message().Payload, ".jpg") || strings.HasSuffix(c.Message().Payload, ".png") || strings.HasSuffix(c.Message().Payload, ".jpeg") {
			toBeSent = &tb.Photo{File: tb.FromURL(c.Message().Payload)}
		} else if strings.HasSuffix(c.Message().Payload, ".gif") {
			toBeSent = &tb.Animation{File: tb.FromURL(c.Message().Payload)}
		} else {
			return nil
		}

		_, err = d.Bot.Send(tb.ChatID(home), toBeSent, &tb.SendOptions{AllowWithoutReply: true})
		if err != nil {
			_, e := d.Bot.Send(c.Message().Chat, "Failed sending that photo: "+err.Error())
			if e != nil {
				return fmt.Errorf("sending message: %w", e)
			}

			return fmt.Errorf("sending photo: %w", err)
		}

		_, err = d.Bot.Send(c.Chat(), "Photo sent")
		if err != nil {
			return fmt.Errorf("sending message that says 'photo sent': %w", err)
		}
		return nil

	}

	_, err = d.Bot.Send(tb.ChatID(home), c.Message().Payload, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		_, e := d.Bot.Send(c.Chat(), "Failed sending that message: "+err.Error())
		if e != nil {
			return fmt.Errorf("sending message: %w", e)
		}

		return fmt.Errorf("sending message: %w", err)
	}

	_, err = d.Bot.Send(c.Chat(), "Message sent")
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	return nil
}
