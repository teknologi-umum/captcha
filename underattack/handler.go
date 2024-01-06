package underattack

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	"github.com/getsentry/sentry-go"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// EnableUnderAttackModeHandler provides a handler for /UnderAttack command.
func (d *Dependency) EnableUnderAttackModeHandler(ctx context.Context, c tb.Context) error {
	if c.Message().Private() || c.Sender().IsBot {
		return nil
	}

	span := sentry.StartSpan(ctx, "bot.enable_under_attack_mode_handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha EnableUnderAttackModeHandler"))
	defer span.Finish()
	ctx = span.Context()

	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "user",
		Category: "command.triggered",
		Message:  "/UnderAttack",
		Data: map[string]interface{}{
			"user": c.Sender(),
			"chat": c.Chat(),
		},
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	admins, err := c.Bot().AdminsOf(ctx, c.Chat())
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	if !utils.IsAdmin(admins, c.Sender()) {
		// It turns out, for contingency reasons, people should be aware that the command and bot
		// is working, yet the bot is rate limited by Telegram because of sending too many messages
		// at one. Hence, we should retry every enabling command for under attack.
		for {
			_, err := c.Bot().Send(
				ctx,
				c.Chat(),
				"Cuma admin yang boleh jalanin command ini. Ada baiknya kamu ping adminnya langsung :)",
				&tb.SendOptions{
					ReplyTo:           c.Message(),
					AllowWithoutReply: true,
				},
			)
			if err != nil {
				if strings.Contains(err.Error(), "retry after") {
					// Acquire the retry number
					retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
					if err != nil {
						// If there's an error, we'll just retry after 15 second
						retry = 15
					}

					// Let's wait a bit and retry
					time.Sleep(time.Second * time.Duration(retry))
					continue
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					time.Sleep(time.Second * 10)
					continue
				}

				shared.HandleBotError(ctx, err, d.Bot, c.Message())
				return nil
			}

			break
		}

		return nil
	}

	// Rate limit per group. Drop if limited
	allowedThrough := RateLimitCall(c.Chat().ID, time.Second*10)
	if !allowedThrough {
		return nil
	}

	// Sender must be an admin here.
	ctx, cancel := context.WithTimeout(ctx, time.Minute*1)
	defer cancel()

	// Check if we are on the under attack mode right now.
	underAttackModeEnabled, err := d.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	if underAttackModeEnabled {
		for {
			_, err := c.Bot().Send(
				ctx,
				c.Chat(),
				"Mode under attack sudah menyala. Untuk mematikan, kirim /disableunderattack",
				&tb.SendOptions{
					ReplyTo:           c.Message(),
					AllowWithoutReply: true,
				},
			)
			if err != nil {
				if strings.Contains(err.Error(), "retry after") {
					// Acquire the retry number
					retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
					if err != nil {
						// If there's an error, we'll just retry after 15 second
						retry = 15
					}

					// Let's wait a bit and retry
					time.Sleep(time.Second * time.Duration(retry))
					continue
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					time.Sleep(time.Second * 10)
					continue
				}

				shared.HandleBotError(ctx, err, d.Bot, c.Message())
				return nil
			}

			break
		}

		return nil
	}

	expiresAt := time.Now().Add(time.Minute * 30)
	var notificationMessage *tb.Message
	for {
		notificationMessage, err = c.Bot().Send(
			ctx,
			c.Chat(),
			"Grup ini dalam kondisi under attack sampai pukul "+
				expiresAt.In(time.FixedZone("WIB", 7*60*60)).Format("15:04 MST")+
				". Semua yang baru masuk ke grup ini akan langsung di ban selamanya. "+
				"Untuk bisa bergabung, tunggu sampai under attack mode berakhir, atau hubungi admin grup.\n\n"+
				"This group is in under attack mode until "+
				expiresAt.In(time.FixedZone("UTC +7", 7*60*60)).Format("15:04 MST")+
				". Everyone that is joining this group will be banned forever. "+
				"To be able to join, wait until the under attack mode is over, or contact the group's administrator.",
			&tb.SendOptions{
				ParseMode: tb.ModeDefault,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "retry after") {
				// Acquire the retry number
				retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
				if err != nil {
					// If there's an error, we'll just retry after 15 second
					retry = 15
				}

				// Let's wait a bit and retry
				time.Sleep(time.Second * time.Duration(retry))
				continue
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				time.Sleep(time.Second * 10)
				continue
			}

			shared.HandleBotError(ctx, err, d.Bot, c.Message())
			return nil
		}

		break
	}

	err = d.Datastore.SetUnderAttackStatus(ctx, c.Chat().ID, true, time.Now().Add(time.Minute*30), int64(notificationMessage.ID))
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	err = d.Memory.Delete("UnderAttack:" + strconv.FormatInt(c.Chat().ID, 10))
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	err = c.Bot().Pin(ctx, notificationMessage)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "debug",
		Category: "underattack.state",
		Message:  "Under attack mode is enabled",
		Data: map[string]interface{}{
			"user": c.Sender(),
			"chat": c.Chat(),
		},
		Level:     sentry.LevelDebug,
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	go func() {
		// Set a timer to unpin the notification message
		time.Sleep(time.Until(expiresAt))

		sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
			Type:     "debug",
			Category: "underattack.state",
			Message:  "Under attack mode ends",
			Data: map[string]interface{}{
				"user": c.Sender(),
				"chat": c.Chat(),
			},
			Level:     sentry.LevelDebug,
			Timestamp: time.Now(),
		}, &sentry.BreadcrumbHint{})

		err := c.Bot().Unpin(ctx, notificationMessage.Chat, notificationMessage.ID)
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, c.Message())
		}
	}()

	return nil
}

// DisableUnderAttackModeHandler provides a handler for /disableunderattack command.
func (d *Dependency) DisableUnderAttackModeHandler(ctx context.Context, c tb.Context) error {
	if c.Message().Private() || c.Sender().IsBot {
		return nil
	}

	span := sentry.StartSpan(ctx, "bot.disable_under_attack_mode_handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha DisableUnderAttackModeHandler"))
	defer span.Finish()
	ctx = span.Context()

	admins, err := c.Bot().AdminsOf(ctx, c.Chat())
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	if !utils.IsAdmin(admins, c.Sender()) {
		_, err := c.Bot().Send(
			ctx,
			c.Chat(),
			"Cuma admin yang boleh jalanin command ini. Ada baiknya kamu ping adminnya langsung :)",
			&tb.SendOptions{
				ReplyTo:           c.Message(),
				AllowWithoutReply: true,
			},
		)
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, c.Message())
			return nil
		}

		return nil
	}

	// Rate limit per group. Drop if limited
	allowedThrough := RateLimitCall(c.Chat().ID, time.Second*10)
	if !allowedThrough {
		return nil
	}

	// Sender must be an admin here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// Check if we are on the under attack mode right now.
	underAttackModeEnabled, err := d.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	if !underAttackModeEnabled {
		return nil
	}

	underAttackEntry, err := d.Datastore.GetUnderAttackEntry(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	err = d.Datastore.SetUnderAttackStatus(ctx, c.Chat().ID, false, time.Now(), 0)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	err = d.Memory.Delete("UnderAttack:" + strconv.FormatInt(c.Chat().ID, 10))
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	err = c.Bot().Unpin(ctx, c.Chat(), int(underAttackEntry.NotificationMessageID))
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, c.Message())
		return nil
	}

	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "debug",
		Category: "underattack.state",
		Message:  "Under attack mode is disabled",
		Data: map[string]interface{}{
			"user": c.Sender(),
			"chat": c.Chat(),
		},
		Level:     sentry.LevelDebug,
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	return nil
}
