package reminder

import (
	"context"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	tb "gopkg.in/telebot.v3"
	"strings"
	"time"
)

func (d *Dependency) Handler(ctx context.Context, c tb.Context) error {
	// Check for user limit, their reminder must not exceed 3
	reminderCount, err := d.CheckUserLimit(ctx, c.Sender().ID)
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)

		err := c.Reply("Sorry, a reminder can't be created because of internal error. Please contact the admin!", &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
		if err != nil {
			return err
		}

		return nil
	}

	if reminderCount >= 3 {
		err := c.Reply("You have exceeded your reminder quota of 3 active reminders per user. Spend your money on a real reminder app.", &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
		if err != nil {
			return err
		}

		return nil
	}

	// Parse text
	reminder, err := ParseText(strings.TrimPrefix(c.Text(), "/remind"))
	if err != nil {
		if errors.Is(err, ErrExceeds24Hours) {
			err := c.Reply("You are attempting to create a reminder that exceeds 24 hour from now. It's prohibited, try shorter time.. or spend your money on a real reminder app.", &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
			if err != nil {
				return err
			}
			return nil
		}

		sentry.GetHubFromContext(ctx).CaptureException(err)

		err := c.Reply("Sorry, I can't create your reminder, something wrong happened on my end. Please contact the admin!", &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
		if err != nil {
			return err
		}
		return nil
	}

	go func(c tb.Context, reminder Reminder) {
		time.Sleep(time.Until(reminder.Time))

		template := fmt.Sprintf("Hi %s! I'm reminding you to %s. Have a great day!", strings.Join(reminder.Subject, ", "), reminder.Object)
		_, err := c.Bot().Send(c.Chat(), template, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
		}
	}(c, reminder)

	return nil
}
