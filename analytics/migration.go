package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
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

	_, err = t.ExecContext(
		ctx,
		`CREATE TABLE analytics (
			user_id INTEGER PRIMARY KEY,
			username VARCHAR(255),
			display_name VARCHAR(255),
			counter INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			joined_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
	)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE INDEX ON analytics (counter)`,
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
