package analytics_test

import (
	"context"
	"database/sql"
	"teknologi-umum-bot/analytics"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestGetUserDataFromDB(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		t.Fatal(err)
	}

	// The lack of group_id value is intentional, because I want to check for
	// null SQL values.
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)`,
		90,
		"user1",
		"User 1",
		1,
		time.Now(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			t.Error(e)
		}
		t.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	deps := &analytics.Dependency{
		DB:     db,
		Memory: memory,
	}

	data, err := deps.GetUserDataFromDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if data[0].UserID != 90 {
		t.Error("user id should be 90, got", data[0].UserID)
	}

	if data[0].Username != "user1" {
		t.Error("username should be user1, got", data[0].Username)
	}

	if data[0].DisplayName != "User 1" {
		t.Error("display name should be User 1, got", data[0].DisplayName)
	}
}

func TestGetHourlyDataFromDB(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// create a dummy hourly type
	hourly := []analytics.HourlyMap{
		{
			TodaysDate: "2021-01-01",
			ZeroHour:   14,
			OneHour:    15,
			TwoHour:    16,
		},
		{
			TodaysDate: "2021-01-02",
			ZeroHour:   3,
			OneHour:    4,
			TwoHour:    5,
		},
		{
			TodaysDate: "2021-01-03",
			ZeroHour:   6,
			OneHour:    7,
			TwoHour:    8,
		},
	}

	// convert hourly slice to a single interface{} slice
	var hourlySlice []interface{}
	for _, v := range hourly {
		hourlySlice = append(hourlySlice, v.TodaysDate, v.ZeroHour, v.OneHour, v.TwoHour)
	}

	c, err := db.Connx(ctx)
	if err != nil {
		t.Error(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		t.Error(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO analytics_hourly
			(todays_date, zero_hour, one_hour, two_hour)
			VALUES
			($1, $2, $3, $4),
			($5, $6, $7, $8),
			($9, $10, $11, $12)`,
		hourlySlice...,
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			t.Error(e)
		}
		t.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			t.Error(e)
		}
		t.Error(err)
	}

	deps := &analytics.Dependency{
		DB:     db,
		Memory: memory,
	}

	data, err := deps.GetHourlyDataFromDB(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 3 {
		t.Error("data length should be 3, got", len(data))
	}
}
