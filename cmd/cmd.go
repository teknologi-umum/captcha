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

func (d *Dependency) OnTextHandler(m *tb.Message) {
	err := d.analytics.NewMsg(m.Sender)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
		return
	}

	d.captcha.WaitForAnswer(m)
}

func (d *Dependency) OnUserJoinHandler(m *tb.Message) {
	var tempSender *tb.User
	if m.UserJoined.ID != 0 {
		tempSender = m.UserJoined
	} else {
		tempSender = m.Sender
	}

	err := d.analytics.NewUser(tempSender)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
		return
	}

	d.captcha.CaptchaUserJoin(m)
}

func (d *Dependency) OnNonTextHandler(m *tb.Message) {
	err := d.analytics.NewMsg(m.Sender)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
		return
	}

	d.captcha.NonTextListener(m)
}
func (d *Dependency) OnUserLeftHandler(m *tb.Message) {
	d.captcha.CaptchaUserLeave(m)
}
func (d *Dependency) AsciiCmdHandler(m *tb.Message) {
	d.ascii.Ascii(m)
}
