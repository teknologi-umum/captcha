package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
)

func (d *Dependency) getUserDataFromDB(ctx context.Context) ([]User, error) {
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

func (d *Dependency) getHourlyDataFromDB(ctx context.Context) ([]Hourly, error) {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return []Hourly{}, nil
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return []Hourly{}, err
	}

	rows, err := tx.QueryxContext(ctx, "SELECT * FROM analytics_hourly")
	if err != nil {
		tx.Rollback()
		return []Hourly{}, err
	}
	defer rows.Close()

	var hourly []Hourly
	for rows.Next() {
		var hour Hourly
		err := rows.StructScan(&hour)
		if err != nil {
			tx.Rollback()
			return []Hourly{}, err
		}

		hourly = append(hourly, hour)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return []Hourly{}, err
	}

	return hourly, nil
}

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
	users, err := d.getUserDataFromDB(ctx)
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

	err = d.Memory.Set("analytics:last_updated:users", []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		return []byte{}, err
	}

	return data, nil
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
	data, err := d.getUserDataFromDB(ctx)
	if err != nil {
		return []byte{}, err
	}

	var tempTotal int
	for _, user := range data {
		tempTotal += user.Counter
	}

	total = []byte(strconv.Itoa(tempTotal))
	err = d.Memory.Set("analytics:total", total)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:last_updated:total", []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		return []byte{}, err
	}

	return total, nil
}

func (d *Dependency) GetHourly(ctx context.Context) ([]byte, error) {
	// get the hourly from the cache first
	hourly, err := d.Memory.Get("analytics:hourly")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return []byte{}, err
	}

	if len(hourly) > 0 {
		return hourly, nil
	}

	// if not in memory cache, then check from the database
	data, err := d.getHourlyDataFromDB(ctx)
	if err != nil {
		return []byte{}, err
	}

	hourly, err = json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:hourly", hourly)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:last_updated:hourly", []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		return []byte{}, err
	}

	return hourly, nil
}

func (d *Dependency) LastUpdated(r int) (time.Time, error) {
	switch r {
	case 0:
		data, err := d.Memory.Get("analytics:last_updated:users")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	case 1:
		data, err := d.Memory.Get("analytics:last_updated:total")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	case 2:
		data, err := d.Memory.Get("analytics:last_updated:hourly")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	default:
		return time.Time{}, errors.New("invalid r value")
	}
}
