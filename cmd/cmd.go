package cmd

import (
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/ascii"
	"teknologi-umum-bot/captcha"
	"teknologi-umum-bot/shared"

	"github.com/allegro/bigcache/v3"
	"github.com/bsm/redislock"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Dependency struct {
	Memory    *bigcache.BigCache
	Redis     *redis.Client
	Locker    *redislock.Client
	Bot       *tb.Bot
	Logger    *sentry.Client
	DB        *sqlx.DB
	captcha   *captcha.Dependencies
	ascii     *ascii.Dependencies
	analytics *analytics.Dependency
}

// New returns a pointer struct of Dependency
// which map the incoming dependencies provided
// into what's needed by each domain.
func New(deps Dependency) *Dependency {
	return &Dependency{
		captcha: &captcha.Dependencies{
			Memory: deps.Memory,
			Redis:  deps.Redis,
			Bot:    deps.Bot,
			Logger: deps.Logger,
		},
		ascii: &ascii.Dependencies{
			Bot: deps.Bot,
		},
		analytics: &analytics.Dependency{
			Memory: deps.Memory,
			Redis:  deps.Redis,
			Locker: deps.Locker,
			Bot:    deps.Bot,
			Logger: deps.Logger,
			DB:     deps.DB,
		},
	}
}

// OnTextHandler handle any incoming text from the group
func (d *Dependency) OnTextHandler(m *tb.Message) {
	err := d.analytics.NewMsg(m)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
		return
	}

	d.captcha.WaitForAnswer(m)
}

// OnUserJoinHandler handle any incoming user join,
// whether they was invited by someone (meaning they are
// added by someone else into the group), or they join
// the group all by themself.
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

	err := d.analytics.NewMsg(m)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
		return
	}
}

// OnUserLeftHandle handles during an event in which
// a user left the group.
func (d *Dependency) OnUserLeftHandler(m *tb.Message) {
	d.captcha.CaptchaUserLeave(m)
}

// AsciiCmdHandler handle the /ascii command.
func (d *Dependency) AsciiCmdHandler(m *tb.Message) {
	d.ascii.Ascii(m)
}
