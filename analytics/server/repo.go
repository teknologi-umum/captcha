package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/allegro/bigcache/v3"
)

func (d *Dependency) GetAll(ctx context.Context) ([]byte, error) {
	// check from in memory cache first for the data
	data, err := d.Memory.Get("analytics:analytics")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return []byte{}, err
	}

	if len(data) > 0 {
		return data, nil
	}

	// if not in memory cache, then check from the database
	users, err := d.getDataFromDB(ctx)
	if err != nil {
		return []byte{}, err
	}

	// convert the struct to json first
	data, err = json.Marshal(users)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:analytics", data)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (d *Dependency) getDataFromDB(ctx context.Context) ([]User, error) {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return []User{}, nil
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return []User{}, err
	}

	rows, err := tx.QueryxContext(ctx, "SELECT * FROM analytics")
	if err != nil {
		tx.Rollback()
		return []User{}, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.StructScan(&user)
		if err != nil {
			tx.Rollback()
			return []User{}, err
		}
		users = append(users, user)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return []User{}, err
	}

	return users, nil
}

func (d *Dependency) GetTotal(ctx context.Context) ([]byte, error) {
	// get the total from the cache first
	total, err := d.Memory.Get("analytics:total")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return []byte{}, err
	}

	if len(total) > 0 {
		return total, nil
	}

	// if not in memory cache, then check from the database
	data, err := d.getDataFromDB(ctx)
	if err != nil {
		return []byte{}, err
	}

	total = []byte(strconv.Itoa(len(data)))
	err = d.Memory.Set("analytics:total", total)
	if err != nil {
		return []byte{}, err
	}

	return total, nil
}
