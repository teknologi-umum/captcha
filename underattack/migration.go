package underattack

import (
	"context"
	"database/sql"
	"errors"
	"teknologi-umum-bot/shared"
	"time"

	"github.com/jmoiron/sqlx"
)

// MustMigrate creates a dependency struct and a context that will execute the Migrate() function
func MustMigrate(db *sqlx.DB) error {
	d := &Dependency{DB: db}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return d.Migrate(ctx)
}

// Migrate will migrates database tables for under attack domain.
func (d *Dependency) Migrate(ctx context.Context) error {
	c, err := d.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			shared.HandleError(ctx, err)
		}
	}()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS under_attack (
			group_id BIGINT PRIMARY KEY,
			is_under_attack BOOLEAN NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			notification_message_id BIGINT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return err
		}

		return err
	}
	_, err = tx.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_updated_at ON under_attack (updated_at)`,
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return err
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return err
		}

		return err
	}

	return nil
}
