package cmd

import (
	"context"
	"strconv"
	"strings"
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/ascii"
	"teknologi-umum-bot/badwords"
	"teknologi-umum-bot/captcha"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/underattack"
	"time"

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
	Logger      *sentry.Client
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
		Logger:   deps.Logger,
		DB:       deps.DB,
		TeknumID: deps.TeknumID,
	}
	return &Dependency{
		Bot:         deps.Bot,
		Logger:      deps.Logger,
		Memory:      deps.Memory,
		DB:          deps.DB,
		Mongo:       deps.Mongo,
		MongoDBName: deps.MongoDBName,
		captcha: &captcha.Dependencies{
			Memory:    deps.Memory,
			Bot:       deps.Bot,
			Logger:    deps.Logger,
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
			Logger: deps.Logger,
		},
	}
}

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(c tb.Context) error {
	d.captcha.WaitForAnswer(c.Message())

	err := d.analytics.NewMessage(c.Message())
	if err != nil {
		shared.HandleError(err, d.Logger)
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

	underAttack, err := d.underAttack.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleError(err, d.Logger)
	}

	if underAttack {
		err := c.Bot().Ban(c.Chat(), &tb.ChatMember{User: c.Sender(), RestrictedUntil: tb.Forever()})
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
			return nil
		}
	}

	var tempSender *tb.User
	if c.Message().UserJoined.ID != 0 {
		tempSender = c.Message().UserJoined
	} else {
		tempSender = c.Message().Sender
	}

	go d.analytics.NewUser(c.Message(), tempSender)

	d.captcha.CaptchaUserJoin(c.Message())

	return nil
}

// OnNonTextHandler meant to handle anything else
// than an incoming text message.
func (d *Dependency) OnNonTextHandler(c tb.Context) error {
	d.captcha.NonTextListener(c.Message())

	err := d.analytics.NewMessage(c.Message())
	if err != nil {
		shared.HandleError(err, d.Logger)
	}

	return nil
}

// OnUserLeftHandler handles during an event in which
// a user left the group.
func (d *Dependency) OnUserLeftHandler(c tb.Context) error {
	d.captcha.CaptchaUserLeave(c.Message())
	return nil
}

// AsciiCmdHandler handle the /ascii command.
func (d *Dependency) AsciiCmdHandler(c tb.Context) error {
	d.ascii.Ascii(c.Message())
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

	err := d.badwords.AddBadWord(ctx, strings.TrimPrefix(c.Message().Text, "/badwords "))
	if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
		shared.HandleBotError(err, d.Logger, c.Bot(), c.Message())
		return nil
	}

	_, err = c.Bot().Send(c.Sender(), "Terimakasih telah menambahkan kata yang tidak pantas.")
	if err != nil {
		shared.HandleBotError(err, d.Logger, c.Bot(), c.Message())
	}

	return nil
}

// CukupHandler was created just to mock laode.
func (d *Dependency) CukupHandler(c tb.Context) error {
	if c.Message().Private() {
		return nil
	}

	_, err := c.Bot().Send(c.Chat(), &tb.Photo{File: tb.FromURL("https://i.ibb.co/WvynnPb/ezgif-4-13e23b17f1.jpg")})
	if err != nil {
		shared.HandleBotError(err, d.Logger, c.Bot(), c.Message())
	}

	return nil
}

// EnableUnderAttackModeHandler provides a handler for /underattack command.
func (d *Dependency) EnableUnderAttackModeHandler(c tb.Context) error {
	return d.underAttack.EnableUnderAttackModeHandler(c)
}

// DisableUnderAttackModeHandler provides a handler for /disableunderattack command.
func (d *Dependency) DisableUnderAttackModeHandler(c tb.Context) error {
	return d.underAttack.DisableUnderAttackModeHandler(c)
}
