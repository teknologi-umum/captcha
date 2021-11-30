package analytics

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependency) NewUser(m *tb.Message, user *tb.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := d.DB.Connx(ctx)
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
	}
	defer c.Close()

	t, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		shared.HandleError(err, d.Logger, d.Bot, m)
	}

	now := time.Now()

	_, err = t.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (user_id)
			DO UPDATE
				SET joined_at = $8,
					updated_at = $9`,
		user.ID,
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
		t.Rollback()
		shared.HandleError(err, d.Logger, d.Bot, m)
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		shared.HandleError(err, d.Logger, d.Bot, m)
	}
}
