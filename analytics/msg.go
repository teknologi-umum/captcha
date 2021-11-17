package analytics

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/tucnak/telebot.v2"
)

func NewMsg(db *sqlx.DB, redis *redis.Client, user *tb.User) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()

	// Check latest hour
	p := redis.TxPipeline()
	defer p.Close()

	hour, err := p.Get(ctx, "analytics:hour").Result()
	if err != nil {
		return err
	}

	now := time.Now().Hour()

	// Create new hour
	if hour == "" && now > time.Now().Hour() {
		counter, err := p.Get(ctx, "analytics:counter").Result()
		if err != nil {
			return err
		}

		// Insert a new counter to Redis, do nothing on the DB
		if counter == "" {

		}

		err = p.Set(ctx, "analytics:hour", strconv.Itoa(now), 0).Err()
		if err != nil {
			return err
		}

	}

	c, err := db.Connx(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	return nil
}
