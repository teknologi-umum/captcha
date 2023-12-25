package main

import (
	"context"
	"fmt"
	"github.com/teknologi-umum/captcha/setir"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/ascii"
	"github.com/teknologi-umum/captcha/badwords"
	"github.com/teknologi-umum/captcha/captcha"
	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/underattack"
	tb "gopkg.in/telebot.v3"
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
	Badwords    *badwords.Dependency
	UnderAttack *underattack.Dependency
	Setir       *setir.Dependency
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

	if deps.FeatureFlag.BadwordsInsertion && deps.Badwords == nil {
		return nil, fmt.Errorf("badwords insertion feature is enabled, but badwords dependency is nil")
	}

	return &deps, nil
}

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

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
	ctx = span.Context()

	if d.FeatureFlag.UnderAttack {
		underAttack, err := d.UnderAttack.AreWe(ctx, c.Chat().ID)
		if err != nil {
			shared.HandleError(ctx, err)
		}

		if underAttack {
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

// BadWordHandler handle the /badwords command.
// This can only be accessed by some users on Telegram
// and only valid for private chats.
func (d *Dependency) BadWordHandler(c tb.Context) error {
	if d.FeatureFlag.BadwordsInsertion {
		return nil
	}

	if !c.Message().Private() {
		return nil
	}
	ok := d.Badwords.Authenticate(strconv.FormatInt(c.Sender().ID, 10))
	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	err := d.Badwords.AddBadWord(ctx, strings.TrimPrefix(c.Message().Text, "/badwords "))
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

func (d *Dependency) SetirHandler(c tb.Context) error {
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())

	return d.Setir.Handler(ctx, c)
}
