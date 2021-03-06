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
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	tb "gopkg.in/tucnak/telebot.v2"
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
			Logger:      deps.Logger,
			Mongo:       deps.Mongo,
			MongoDBName: deps.MongoDBName,
		},
	}
}

var globalMsgs = make(chan *tb.Message)

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(m *tb.Message) {
	d.captcha.WaitForAnswer(m)

	err := d.analytics.NewMessage(m)
	if err != nil {
		shared.HandleError(err, d.Logger)
		return
	}
}

// OnUserJoinHandler handle any incoming user join,
// whether they were invited by someone (meaning they are
// added by someone else into the group), or they join
// the group all by themselves.
func (d *Dependency) OnUserJoinHandler(m *tb.Message) {
	var tempSender *tb.User
	if m.UserJoined.ID != 0 {
		tempSender = m.UserJoined
	} else {
		tempSender = m.Sender
	}

	go d.analytics.NewUser(m, tempSender)

	d.captcha.CaptchaUserJoin(m)
}

// OnNonTextHandler meant to handle anything else
// than an incoming text message.
func (d *Dependency) OnNonTextHandler(m *tb.Message) {
	d.captcha.NonTextListener(m)

	err := d.analytics.NewMessage(m)
	if err != nil {
		shared.HandleError(err, d.Logger)
		return
	}
}

// OnUserLeftHandler handles during an event in which
// a user left the group.
func (d *Dependency) OnUserLeftHandler(m *tb.Message) {
	d.captcha.CaptchaUserLeave(m)
}

// AsciiCmdHandler handle the /ascii command.
func (d *Dependency) AsciiCmdHandler(m *tb.Message) {
	d.ascii.Ascii(m)
}

// BadWordsCmdHandler handle the /badwords command.
// This can only be accessed by some users on Telegram
// and only valid for private chats.
func (d *Dependency) BadWordHandler(m *tb.Message) {
	if !m.Private() {
		return
	}
	ok := d.badwords.Authenticate(strconv.FormatInt(m.Sender.ID, 10))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err := d.badwords.AddBadWord(ctx, strings.TrimPrefix(m.Text, "/badwords "))
	if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	_, err = d.Bot.Send(m.Sender, "Terimakasih telah menambahkan kata yang tidak pantas.")
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}
}

// CukupHandler was created just to mock laode.
func (d *Dependency) CukupHandler(m *tb.Message) {
	if m.Private() {
		return
	}

	_, err := d.Bot.Send(m.Chat, &tb.Photo{File: tb.FromURL("https://i.ibb.co/WvynnPb/ezgif-4-13e23b17f1.jpg")})
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}
}
