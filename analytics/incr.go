package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"teknologi-umum-bot/shared"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

// IncrementUserDB literally increment a user's counter on the database.
func (d *Dependency) IncrementUserDB(ctx context.Context, member GroupMember) error {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			shared.HandleError(err, d.Logger)
		}
	}(c)

	t, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: false})
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = t.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at, group_id)
			VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id)
			DO UPDATE
			SET counter = (SELECT counter FROM analytics WHERE user_id = $1)+$4,
				username = $2,
				display_name = $3,
				updated_at = $7`,
		member.UserID,
		member.Username,
		member.DisplayName,
		member.Counter,
		now,
		now,
		now,
		member.GroupID,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}
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
		fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day()),
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}
		return err
	}

	err = t.Commit()
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}
		return errors.Wrap(err, "failed on incrementing counter")
	}

	return nil
}
