package reminder

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/utils"
	tb "gopkg.in/telebot.v3"
)

func (d *Dependency) Handler(ctx context.Context, c tb.Context) error {
	input := strings.TrimPrefix(strings.TrimPrefix(c.Text(), "/remind@TeknumCaptchaBot"), "/remind")
	if input == "" {
		err := c.Reply(
			"To use /remind properly, you should add with the remaining text including the normal human grammatical that I can understand.\n\n"+
				"For English, see this <a href=\"https://www.grammarly.com/blog/sentence-structure/\">Grammarly article about sentence structure</a>.\n"+
				"Untuk Indonesia, pakai SPO + Keterangan Waktu yang baik dan benar, <a href=\"https://tambahpinter.com/bentuk-kalimat-spok/\">belajar lagi biar tambah pinter</a>.",
			&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
		)
		if err != nil {
			return err
		}

		return nil
	}

	// Check for user limit, their reminder must not exceed 3
	reminderCount, err := d.CheckUserLimit(ctx, c.Sender().ID)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)

		err := c.Reply(
			"Sorry, a reminder can't be created because of internal error. Please contact the admin!",
			&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
		)
		if err != nil {
			return err
		}

		return nil
	}

	if reminderCount >= 3 {
		err := c.Reply(
			"You have exceeded your reminder quota of 3 active reminders per user. Spend your money on a real reminder app.",
			&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
		)
		if err != nil {
			return err
		}

		return nil
	}

	// This is an experimental feature, so sending telemetry is a must.
	// It eases the debugging process.
	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "info",
		Category: "reminder.handler",
		Message:  "A reminder input just came in",
		Data: map[string]interface{}{
			"Reminder Text":    input,
			"Chat ID":          c.Chat().ID,
			"Chat Username":    c.Chat().Username,
			"Chat Full Name":   c.Chat().FirstName + " " + c.Chat().LastName,
			"Chat Title":       c.Chat().Title,
			"Message ID":       c.Message().ID,
			"Sender ID":        c.Sender().ID,
			"Sender Username":  c.Sender().Username,
			"Sender Full Name": c.Sender().FirstName + " " + c.Sender().LastName,
			"From Group":       c.Message().FromGroup(),
			"From Channel":     c.Message().FromChannel(),
			"Is Forwarded":     c.Message().IsForwarded(),
		},
		Level:     "debug",
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	// Parse text
	reminder, err := ParseText(ctx, input)
	if err != nil {
		if errors.Is(err, ErrExceeds24Hours) {
			err := c.Reply(
				"You are attempting to create a reminder that exceeds 24 hour from now. It's prohibited, try shorter time.. or spend your money on a real reminder app.",
				&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
			)
			if err != nil {
				return err
			}
			return nil
		}

		sentry.GetHubFromContext(ctx).CaptureException(err)

		err := c.Reply(
			"Sorry, I can't create your reminder, something wrong happened on my end. Please contact the admin!",
			&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
		)
		if err != nil {
			return err
		}
		return nil
	}

	if reminder.Time.IsZero() || len(reminder.Subject) == 0 || reminder.Object == "" || reminder.Time.Unix() < time.Now().Unix() {
		err := c.Reply(
			"Sorry, I'm unable to parse the reminder text that you just sent. Send /remind and see the guide for this command.",
			&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
		)
		if err != nil {
			return err
		}

		return nil
	}

	for i, subject := range reminder.Subject {
		if subject == "me" {
			reminder.Subject[i] = "<a href=\"tg://user?id=" + strconv.FormatInt(c.Message().Sender.ID, 10) + "\">" +
				utils.SanitizeInput(c.Message().Sender.FirstName) + utils.ShouldAddSpace(c.Message().Sender) + utils.SanitizeInput(c.Message().Sender.LastName) +
				"</a>"
			break
		}
	}

	// Start the reminder goroutine
	go func(c tb.Context, reminder Reminder) {
		time.Sleep(time.Until(reminder.Time))

		template := fmt.Sprintf(
			"Hi %s! I'm reminding you to %s. Have a great day!",
			strings.Join(reminder.Subject, ", "),
			utils.SanitizeInput(reminder.Object),
		)
		_, err := c.Bot().Send(c.Chat(), template, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
		}
	}(c, reminder)

	err = d.IncrementUserLimit(ctx, c.Sender().ID, reminderCount+1)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
	}

	err = c.Reply(fmt.Sprintf("Reminder for %s was created", reminder.Time.Format(time.RFC1123)))
	if err != nil {
		return err
	}

	return nil
}
