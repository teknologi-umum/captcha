package analytics

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/shared"

	"github.com/jmoiron/sqlx"
)

// Returns a slice of GroupMember from the database.
func (d *Dependency) GetUserDataFromDB(ctx context.Context) ([]GroupMember, error) {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return []GroupMember{}, nil
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			shared.HandleError(err, d.Logger)
		}
	}(c)

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return []GroupMember{}, err
	}

	rows, err := tx.QueryxContext(ctx, "SELECT * FROM analytics")
	if err != nil {
		if r := tx.Rollback(); r != nil {
			return []GroupMember{}, err
		}
		return []GroupMember{}, err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			shared.HandleError(err, d.Logger)
		}
	}(rows)

	var users []GroupMember
	for rows.Next() {
		var user GroupMember
		err := rows.StructScan(&user)
		if err != nil {
			if r := tx.Rollback(); r != nil {
				return []GroupMember{}, err
			}
			return []GroupMember{}, err
		}
		users = append(users, user)
	}

	err = tx.Commit()
	if err != nil {
		if r := tx.Rollback(); r != nil {
			return []GroupMember{}, err
		}
		return []GroupMember{}, err
	}

	return users, nil
}

// Return a slice of HourlyMap from the database.
func (d *Dependency) GetHourlyDataFromDB(ctx context.Context) ([]HourlyMap, error) {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return []HourlyMap{}, nil
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			shared.HandleError(err, d.Logger)
		}
	}(c)

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return []HourlyMap{}, err
	}

	rows, err := tx.QueryxContext(ctx, "SELECT * FROM analytics_hourly")
	if err != nil {
		if r := tx.Rollback(); r != nil {
			return []HourlyMap{}, nil
		}
		return []HourlyMap{}, err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			shared.HandleError(err, d.Logger)
		}
	}(rows)

	var hourly []HourlyMap
	for rows.Next() {
		var hour HourlyMap
		err := rows.StructScan(&hour)
		if err != nil {
			if r := tx.Rollback(); r != nil {
				return []HourlyMap{}, nil
			}
			return []HourlyMap{}, err
		}

		hourly = append(hourly, hour)
	}

	err = tx.Commit()
	if err != nil {
		if r := tx.Rollback(); r != nil {
			return []HourlyMap{}, nil
		}
		return []HourlyMap{}, err
	}

	return hourly, nil
}
