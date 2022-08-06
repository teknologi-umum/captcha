package analytics

import (
	"context"
	"strconv"
	"time"

	tb "gopkg.in/telebot.v3"
)

// NewMessage handles an incoming message from the group
// to be noted into the database.
func (d *Dependency) NewMessage(m *tb.Message) error {
	// fast return if it's not from a group
	if !m.FromGroup() {
		return nil
	}

	if strconv.FormatInt(m.Chat.ID, 10) != d.TeknumID {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usr := ParseGroupMember(m)
	usr.Counter = 1

	err := d.IncrementUserDB(ctx, usr)
	if err != nil {
		return err
	}

	return nil
}
