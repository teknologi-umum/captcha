package analytics

import (
	"context"
	"database/sql"
	"strconv"
	"time"
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
		r, err := t.QueryxContext(
			ctx,
			`SELECT counter FROM analytics WHERE user_id = $1`,
			user.UserID,
		)
		if err != nil {
			t.Rollback()
			return err
		}
		defer r.Close()

		var counter int
		if r.Next() {
			err = r.Scan(&counter)
			if err != nil {
				t.Rollback()
				return err
			}
		}

		now := time.Now()

		_, err = t.ExecContext(
			ctx,
			`UPDATE analytics
				SET counter = $1,
					updated_at = $2,
					username = $3,
					display_name = $4
				WHERE user_id = $5`,
			counter+user.Counter,
			now,
			now,
			user.Username,
			user.DisplayName,
			user.UserID,
		)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		return err
	}

	return nil
}

func (d *Dependency) IncrementUsrRedis(ctx context.Context, user UserMap) error {
	p := d.Redis.TxPipeline()
	defer p.Close()

	// Per Redis' documentation, INCR will create a new key
	// if the named key does not exists in the first place.
	err := p.Incr(ctx, "analytics:"+strconv.FormatInt(user.UserID, 10)).Err()
	if err != nil {
		return err
	}

	err = p.Do(ctx).Err()
	if err != nil {
		return err
	}

	return nil
}