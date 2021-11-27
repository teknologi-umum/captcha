package analytics

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/aldy505/decrr"
	"github.com/go-redis/redis/v8"
)

func (d *Dependency) IncrementUsrDB(ctx context.Context, users []UserMap) error {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	t, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	for _, user := range users {
		now := time.Now()

		_, err = t.ExecContext(
			ctx,
			`INSERT INTO analytics
				(user_id, username, display_name, counter, created_at, joined_at, updated_at)
				VALUES
				($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (user_id)
				DO UPDATE
				SET counter = (SELECT counter FROM analytics WHERE user_id = $1)+$4,
					username = $2,
					display_name = $3,
					updated_at = $7`,
			user.UserID,
			user.Username,
			user.DisplayName,
			user.Counter,
			now,
			now,
			now,
		)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		return decrr.Wrap(err)
	}

	return nil
}

func (d *Dependency) IncrementUsrRedis(ctx context.Context, user UserMap) error {
	p := d.Redis.TxPipeline()
	defer p.Close()

	usrID := strconv.FormatInt(user.UserID, 10)

	exists, err := d.Redis.HExists(ctx, "analytics:"+usrID, "count").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return decrr.Wrap(err)
	}

	if !exists {
		p.HSet(
			ctx,
			"analytics:"+usrID,
			"counter",
			0,
			"username",
			user.Username,
			"display_name",
			user.DisplayName,
		)
	}

	// Per Redis' documentation, INCR will create a new key
	// if the named key does not exists in the first place.
	p.HIncrBy(ctx, "analytics:"+usrID, "counter", 1)

	// Add the user ID into the Sets of users
	p.SAdd(ctx, "analytics:users", usrID)

	_, err = p.Exec(ctx)
	if err != nil {
		return decrr.Wrap(err)
	}

	return nil
}
