package analytics

import (
	"context"
	"teknologi-umum-bot/shared"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependency) NewMsg(m *tb.Message) error {
	user := m.Sender

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usr := ParseToUser(user)
	usr.Counter = 1

	err := d.IncrementUsrDB(ctx, usr)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
	}

	return nil
}
