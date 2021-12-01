package analytics

import (
	"context"
	"errors"
	"strconv"
	"teknologi-umum-bot/shared"
	"time"

	"github.com/aldy505/decrr"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependency) NewMsg(m *tb.Message) error {
	user := m.Sender

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usr := ParseToUser(user)

	// Whatever we do, we must always increment
	// the user's counter.
	err := d.IncrementUsrRedis(ctx, usr)
	if err != nil {
		return err
	}

	// Check latest hour
	hour, err := d.Redis.Get(ctx, "analytics:hour").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return decrr.Wrap(err)
	}

	now := time.Now().Hour()

	// Create new hour
	if hour == "" {
		err = d.Redis.Set(ctx, "analytics:hour", strconv.Itoa(now), 0).Err()
		if err != nil {
			return decrr.Wrap(err)
		}

		return nil
	}

	hourInt, err := strconv.Atoi(hour)
	if err != nil {
		return err
	}

	// If hourInt < now, insert the data to the database
	if hourInt < now {
		// Create new context
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute*3)
		defer cancel()

		userMaps, err := d.GetAllUserMap(ctx)
		if err != nil {
			shared.HandleError(err, d.Logger, d.Bot, m)
		}

		err = d.IncrementUsrDB(ctx, userMaps)
		if err != nil {
			shared.HandleError(err, d.Logger, d.Bot, m)
		}

		err = d.FlushAllUserID(ctx)
		if err != nil {
			shared.HandleError(err, d.Logger, d.Bot, m)
		}

		err = d.Redis.Set(ctx, "analytics:hour", strconv.Itoa(now), 0).Err()
		if err != nil {
			shared.HandleError(err, d.Logger, d.Bot, m)
		}

		return nil
	}

	return nil
}
