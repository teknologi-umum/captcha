package setir

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

type Dependency struct {
	AdminIDs []string
	HomeID   int64
	Bot      *tb.Bot
}

func New(bot *tb.Bot, adminIDs []string, homeID int64) (*Dependency, error) {
	if bot == nil {
		return nil, fmt.Errorf("bot is nil")
	}

	return &Dependency{
		AdminIDs: adminIDs,
		HomeID:   homeID,
		Bot:      bot,
	}, nil
}

func (d *Dependency) Handler(ctx context.Context, c tb.Context) (err error) {
	if d.AdminIDs == nil || d.HomeID == 0 {
		// The feature is disabled
		return nil
	}

	if !slices.Contains(d.AdminIDs, strconv.FormatInt(c.Sender().ID, 10)) {
		return nil
	}

	if !c.Message().Private() {
		return nil
	}

	span := sentry.StartSpan(ctx, "setir.handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithDescription("Setir Handler"))
	defer span.Finish()
	ctx = span.Context()

	if c.Message().IsReply() {
		var replyToID int

		if strings.HasPrefix(c.Message().Payload, "https://t.me/") {
			replyToID, err = strconv.Atoi(strings.Split(c.Message().Payload, "/")[4])
			if err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				return nil
			}
		} else {
			replyToID, err = strconv.Atoi(c.Message().Payload)
			if err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				return nil
			}
		}

		_, err = d.Bot.Send(ctx, tb.ChatID(d.HomeID), c.Message().ReplyTo.Text, &tb.SendOptions{
			ParseMode:         tb.ModeHTML,
			AllowWithoutReply: true,
			ReplyTo: &tb.Message{
				ID: replyToID,
				Chat: &tb.Chat{
					ID: d.HomeID,
				},
			},
		})
		if err != nil {
			_, err = d.Bot.Send(ctx, c.Chat(), "Failed sending that message: "+err.Error())
			if err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("failed sending that message: %w", err))
				return nil
			}
		} else {
			_, err = d.Bot.Send(ctx, c.Chat(), "Message sent")
			if err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending message: %w", err))
				return nil

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

		_, err = d.Bot.Send(ctx, tb.ChatID(d.HomeID), toBeSent, &tb.SendOptions{AllowWithoutReply: true})
		if err != nil {
			_, e := d.Bot.Send(ctx, c.Message().Chat, "Failed sending that photo: "+err.Error())
			if e != nil {
				sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending message: %w", e))
				return nil
			}

			sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending photo: %w", err))
			return nil
		}

		_, err = d.Bot.Send(ctx, c.Chat(), "Photo sent")
		if err != nil {
			return fmt.Errorf("sending message that says 'photo sent': %w", err)
		}
		return nil

	}

	_, err = d.Bot.Send(ctx, tb.ChatID(d.HomeID), c.Message().Payload, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		_, e := d.Bot.Send(ctx, c.Chat(), "Failed sending that message: "+err.Error())
		if e != nil {
			sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending message: %w", e))
			return nil
		}

		sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending message: %w", err))
		return nil
	}

	_, err = d.Bot.Send(ctx, c.Chat(), "Message sent")
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(fmt.Errorf("sending message: %w", err))
		return nil
	}

	return nil
}
