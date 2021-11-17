package analytics

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/utils"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewUser(db *sqlx.DB, redis *redis.Client, user *tb.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	t, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = t.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)
			ON DUPLICATE KEY
			UPDATE
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
		return err
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		return err
	}

	return nil
}
