package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/teknologi-umum/captcha/deletion"
	"github.com/teknologi-umum/captcha/internal/requestid"
	"github.com/teknologi-umum/captcha/reminder"
	"github.com/teknologi-umum/captcha/setir"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/ascii"
	"github.com/teknologi-umum/captcha/captcha"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/underattack"
)

// Dependency contains the dependency injection struct
// that is required for the main command to use.
//
// It will spread and use the correct dependencies for
// each packages on the captcha project.
type Dependency struct {
	FeatureFlag FeatureFlag
	Captcha     *captcha.Dependencies
	Ascii       *ascii.Dependencies
	Analytics   *analytics.Dependency
	UnderAttack *underattack.Dependency
	Setir       *setir.Dependency
	Reminder    *reminder.Dependency
	Deletion    *deletion.Dependency
}

// New returns a pointer struct of Dependency
// which map the incoming dependencies provided
// into what's needed by each domain.
func New(deps Dependency) (*Dependency, error) {
	// Validate dependencies
	if deps.Captcha == nil {
		return nil, fmt.Errorf("captcha dependency is nil")
	}

	if deps.FeatureFlag.UnderAttack && deps.UnderAttack == nil {
		return nil, fmt.Errorf("under attack feature is enabled, but underattack dependency is nil")
	}

	if deps.FeatureFlag.Analytics && deps.Analytics == nil {
		return nil, fmt.Errorf("analytics feature is enabled, but analytics dependency is nil")
	}

	if deps.FeatureFlag.Reminder && deps.Reminder == nil {
		return nil, fmt.Errorf("reminder feature is enabled, but reminder dependency is nil")
	}

	return &deps, nil
}

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(c tb.Context) error {
	ctx := requestid.SetRequestIdOnContext(sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone()))

	d.Captcha.WaitForAnswer(ctx, c.Message())

	if d.FeatureFlag.Analytics {
		err := d.Analytics.NewMessage(c.Message())
		if err != nil {
			shared.HandleError(ctx, err)
		}
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

	span := sentry.StartSpan(ctx, "bot.on_user_join_handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha OnUserJoinHandler"))
	defer span.Finish()
	ctx = requestid.SetRequestIdOnContext(span.Context())

	if d.FeatureFlag.UnderAttack {
		underAttack, err := d.UnderAttack.AreWe(ctx, c.Chat().ID)
		if err != nil {
			shared.HandleError(ctx, err)
		}

		if underAttack {
			slog.DebugContext(ctx, "State is on under attack mode, preventing a user to come through", requestid.GetSlogAttributesFromContext(ctx)...)
			err := d.UnderAttack.Kicker(ctx, c)
			if err != nil {
				shared.HandleBotError(ctx, err, c.Bot(), c.Message())
			}
			return nil
		}
	}

	var tempSender *tb.User
	if c.Message().UserJoined.ID != 0 {
		tempSender = c.Message().UserJoined
	} else {
		tempSender = c.Message().Sender
	}

	if d.FeatureFlag.Analytics {
		go d.Analytics.NewUser(ctx, c.Message(), tempSender)
	}

	slog.DebugContext(ctx, "Presenting a captcha challenge to the user", slog.String("user_name", tempSender.Username), slog.Int64("user_id", tempSender.ID))
	d.Captcha.CaptchaUserJoin(ctx, c.Message())

	return nil
}

// OnNonTextHandler meant to handle anything else
// than an incoming text message.
func (d *Dependency) OnNonTextHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.Captcha.NonTextListener(ctx, c.Message())

	if d.FeatureFlag.Analytics {
		err := d.Analytics.NewMessage(c.Message())
		if err != nil {
			shared.HandleError(ctx, err)
		}
	}

	return nil
}

// OnUserLeftHandler handles during an event in which
// a user left the group.
func (d *Dependency) OnUserLeftHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	d.Captcha.CaptchaUserLeave(ctx, c.Message())
	return nil
}

// EnableUnderAttackModeHandler provides a handler for /underattack command.
func (d *Dependency) EnableUnderAttackModeHandler(c tb.Context) error {
	if !d.FeatureFlag.UnderAttack {
		return nil
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.UnderAttack.EnableUnderAttackModeHandler(ctx, c)
}

// DisableUnderAttackModeHandler provides a handler for /disableunderattack command.
func (d *Dependency) DisableUnderAttackModeHandler(c tb.Context) error {
	if !d.FeatureFlag.UnderAttack {
		return nil
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.UnderAttack.DisableUnderAttackModeHandler(ctx, c)
}

func (d *Dependency) ReminderHandler(c tb.Context) error {
	if !d.FeatureFlag.Reminder {
		return nil
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	span := sentry.StartSpan(ctx, "bot.reminder_handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha ReminderHandler"))
	defer span.Finish()
	ctx = span.Context()

	return d.Reminder.Handler(ctx, c)
}

func (d *Dependency) SetirHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.Setir.Handler(ctx, c)
}

func (d *Dependency) DeletionHandler(c tb.Context) error {
	if !d.FeatureFlag.Deletion {
		return nil
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())
	span := sentry.StartSpan(ctx, "bot.deletion_handler", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha DeletionHandler"))
	defer span.Finish()
	ctx = span.Context()

	return d.Deletion.Handler(ctx, c)
}
