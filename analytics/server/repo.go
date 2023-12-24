package server

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/dukun"

	"github.com/allegro/bigcache/v3"
)

// GetAll fetch the users' data either it's from the in memory cache
// or from the database (if the data does not exist on the memory cache).
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
	users, err := (&analytics.Dependency{DB: d.DB}).GetUserDataFromDB(ctx)
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

// GetTotal returns the total amount of chat as the data have it.
// The data is fetched from the memory cache or the database,
// if the memory cache data is empty or non-existent.
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
	data, err := (&analytics.Dependency{DB: d.DB}).GetUserDataFromDB(ctx)
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

// GetHourly returns hourly message count, specified in a daily kind of object.
// The data is fetched from the memory cache or the database,
// if the memory cache data is empty or non-existent.
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
	data, err := (&analytics.Dependency{DB: d.DB}).GetHourlyDataFromDB(ctx)
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

// LastUpdated returns the last updated value as a time.Time object
// for cached data.
func (d *Dependency) LastUpdated(r Endpoint) (time.Time, error) {
	switch r {
	case UserEndpoint:
		data, err := d.Memory.Get("analytics:last_updated:users")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	case TotalEndpoint:
		data, err := d.Memory.Get("analytics:last_updated:total")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	case HourlyEndpoint:
		data, err := d.Memory.Get("analytics:last_updated:hourly")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	case DukunEndpoint:
		data, err := d.Memory.Get("analytics:last_updated:dukun")
		if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
			return time.Time{}, err
		}

		if len(data) > 0 {
			return time.Parse(time.RFC3339, string(data))
		}

		return time.Time{}, nil
	default:
		return time.Time{}, ErrInvalidValue
	}
}

func (d *Dependency) GetDukunPoints(ctx context.Context) ([]byte, error) {
	// get the dukun points from the cache first
	dukunPoints, err := d.Memory.Get("analytics:dukun")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return []byte{}, err
	}

	if len(dukunPoints) > 0 {
		return dukunPoints, nil
	}

	// if not in memory cache, then check from the database
	data, err := (&dukun.Dependency{DBName: d.MongoDBName, Mongo: d.Mongo}).GetAllDukun(ctx)
	if err != nil {
		return []byte{}, err
	}

	dukunPoints, err = json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:dukun", dukunPoints)
	if err != nil {
		return []byte{}, err
	}

	err = d.Memory.Set("analytics:last_updated:dukun", []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		return []byte{}, err
	}

	return dukunPoints, nil
}
