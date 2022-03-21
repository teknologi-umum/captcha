package analytics

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/shared"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// MustMigrate is the same as Migrate, but you don't
// need to explicitly create a Dependency struct
// instance. Just supply the database, and you're good
// to go. It will not panic on error, instead it will
// just return an error.
func MustMigrate(db *sqlx.DB) error {
	d := &Dependency{
		DB: db,
	}

	return d.Migrate()
}

// Migrate creates a migration to the database.
// This can be called multiple times as it uses PostgreSQL
// syntax of `IF NOT EXISTS`.
func (d *Dependency) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

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

	t, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS captcha_swarm (
			user_id BIGINT NOT NULL,
			group_id BIGINT NOT NULL,
			username VARCHAR(255),
			display_name VARCHAR(255),
			finished_captcha BOOLEAN NOT NULL,
			joined_at TIMESTAMP
		)`,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}

		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS analytics (
			user_id 		BIGINT	 		PRIMARY KEY,
			group_id 		BIGINT,
			username 		VARCHAR(255),
			display_name 	VARCHAR(255),
			counter 		INTEGER 		DEFAULT 0,
			created_at 		TIMESTAMP 		DEFAULT CURRENT_TIMESTAMP,
			joined_at 		TIMESTAMP,
			updated_at 		TIMESTAMP
		)`,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}

		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_counter ON analytics (counter)`,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}

		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_active ON analytics (updated_at)`,
	)
	if err != nil {
		if r := t.Rollback(); r != nil {
			return r
		}

		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS analytics_hourly (
			todays_date 		VARCHAR(20)	UNIQUE,
			zero_hour 			INTEGER 	DEFAULT 0,
			one_hour 			INTEGER 	DEFAULT 0,
			two_hour 			INTEGER 	DEFAULT 0,
			three_hour 			INTEGER 	DEFAULT 0,
			four_hour 			INTEGER 	DEFAULT 0,
			five_hour 			INTEGER 	DEFAULT 0,
			six_hour 			INTEGER 	DEFAULT 0,
			seven_hour 			INTEGER 	DEFAULT 0,
			eight_hour 			INTEGER 	DEFAULT 0,
			nine_hour 			INTEGER 	DEFAULT 0,
			ten_hour 			INTEGER 	DEFAULT 0,
			eleven_hour 		INTEGER 	DEFAULT 0,
			twelve_hour 		INTEGER 	DEFAULT 0,
			thirteen_hour 		INTEGER 	DEFAULT 0,
			fourteen_hour 		INTEGER 	DEFAULT 0,
			fifteen_hour 		INTEGER 	DEFAULT 0,
			sixteen_hour 		INTEGER 	DEFAULT 0,
			seventeen_hour 		INTEGER 	DEFAULT 0,
			eighteen_hour 		INTEGER 	DEFAULT 0,
			nineteen_hour 		INTEGER 	DEFAULT 0,
			twenty_hour 		INTEGER 	DEFAULT 0,
			twentyone_hour 		INTEGER 	DEFAULT 0,
			twentytwo_hour 		INTEGER 	DEFAULT 0,
			twentythree_hour 	INTEGER 	DEFAULT 0
		)`,
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

		return err
	}

	return nil
}
