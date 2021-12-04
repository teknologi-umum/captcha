package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/aldy505/decrr"
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
