package analytics

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
	"time"

	"github.com/jmoiron/sqlx"

	tb "gopkg.in/tucnak/telebot.v2"
)

// NewUser adds a newly joined user on the group into the database.
//
// If the user has joined before, meaning he left the group for some
// reason, their data should still be here. But, their joined date
// will be updated to their newest join date.
func (d *Dependency) NewUser(m *tb.Message, user *tb.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := d.DB.Connx(ctx)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			shared.HandleError(err, d.Logger)
		}
	}(c)

	t, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
	}

	now := time.Now()

	_, err = t.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, group_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id)
			DO UPDATE
				SET joined_at = $9,
					updated_at = $10`,
		user.ID,
		m.Chat.ID,
		user.Username,
		user.FirstName+utils.ShouldAddSpace(user)+user.LastName,
		0,
		now,
		now,
		now,
		now,
		now,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			shared.HandleError(r, d.Logger)
		}
		shared.HandleBotError(err, d.Logger, d.Bot, m)
	}

	err = t.Commit()
	if err != nil {
		if r := t.Rollback(); r != nil {
			shared.HandleError(r, d.Logger)
		}
		shared.HandleBotError(err, d.Logger, d.Bot, m)
	}
}
