package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

func MustMigrate(db *sqlx.DB) error {
	d := &Dependency{
		DB: db,
	}

	return d.Migrate()
}

func (d *Dependency) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := d.DB.Connx(ctx)
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
		`CREATE TABLE IF NOT EXISTS analytics (
			user_id 		INTEGER 		PRIMARY KEY,
			username 		VARCHAR(255),
			display_name 	VARCHAR(255),
			counter 		INTEGER 		DEFAULT 0,
			created_at 		TIMESTAMP 		DEFAULT CURRENT_TIMESTAMP,
			joined_at 		TIMESTAMP,
			updated_at 		TIMESTAMP
		)`,
	)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_counter ON analytics (counter)`,
	)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS analytics_hourly (
			todays_date 		TIMESTAMP 	PRIMARY KEY,
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
