package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func (d *Dependency) IncrementUsrDB(ctx context.Context, user UserMap) error {
	c, err := d.DB.Connx(ctx)
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

	hourlyQuery := fmt.Sprintf(
		`INSERT INTO analytics_hourly
		(todays_date, %s)
		VALUES
		($1, 1)
		ON CONFLICT (todays_date) DO UPDATE
		SET %s = (SELECT %s FROM analytics_hourly WHERE todays_date = $1)+1`,
		HourMapper[now.Hour()],
		HourMapper[now.Hour()],
		HourMapper[now.Hour()],
	)

	_, err = t.ExecContext(
		ctx,
		hourlyQuery,
		now,
	)
	if err != nil {
		t.Rollback()
		return err
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		return errors.Wrap(err, "failed on incrementing counter")
	}

	return nil
}
