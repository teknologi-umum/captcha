package analytics

import (
	"context"
	"strconv"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependency) NewMsg(user *tb.User) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()

	usr := ParseToUser(user)

	// Check latest hour
	hour, err := d.Redis.Get(ctx, "analytics:hour").Result()
	if err != nil {
		return err
	}

	now := time.Now().Hour()

	// Create new hour
	if hour == "" || now > time.Now().Hour() {
		counter, err := d.Redis.Get(ctx, "analytics:counter").Result()
		if err != nil {
			return err
		}

		// Insert a new counter to Redis, do nothing on the DB
		if counter == "" {
			err = d.IncrementUsrRedis(ctx, usr)
			if err != nil {
				return err
			}
			return nil
		}

		err = d.Redis.Set(ctx, "analytics:hour", strconv.Itoa(now), 0).Err()
		if err != nil {
			return err
		}

		return nil
	}

	// If current hour = hour on redis
	err = d.IncrementUsrRedis(ctx, usr)
	if err != nil {
		return err
	}
	return nil
}
