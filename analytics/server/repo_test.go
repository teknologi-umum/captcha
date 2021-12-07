package server_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"teknologi-umum-bot/analytics/server"
	"testing"
	"time"
)

func TestGetAll(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

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

	deps := &server.Dependency{
		DB:     db,
		Memory: memory,
	}

	res, err := deps.GetAll(ctx)
	if err != nil {
		t.Error(err)
	}

	var user []server.User
	err = json.Unmarshal(res, &user)
	if err != nil {
		t.Error(err)
	}

	if len(user) != 1 {
		t.Errorf("Expected 1 user, got %d", len(user))
	}

	if user[0].UserID != 90 {
		t.Error("user id should be 90, got:", user[0].UserID)
	}

	if user[0].Username != "user1" {
		t.Error("username should be user1, got:", user[0].Username)
	}

	if user[0].DisplayName != "User 1" {
		t.Error("display name should be User 1, got:", user[0].DisplayName)
	}

	if user[0].Counter != 1 {
		t.Error("counter should be 1, got:", user[0].Counter)
	}

	// try to get it again
	res2, err := deps.GetAll(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(res) != string(res2) {
		t.Errorf("result should be the same\n\nres1: %s\n\nres2: %s", string(res), string(res2))
	}
}

func TestGetTotal(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// create a dummy user struct slice
	users := []server.User{
		{UserID: 100, Username: "user1", DisplayName: "User 1", Counter: 1},
		{UserID: 200, Username: "user2", DisplayName: "User 2", Counter: 2},
		{UserID: 300, Username: "user3", DisplayName: "User 3", Counter: 3},
	}

	// convert users slice to single slice with no keys, just values.
	var usersSlice []interface{}
	for _, v := range users {
		usersSlice = append(
			usersSlice,
			v.UserID,
			v.Username,
			v.DisplayName,
			v.Counter,
			time.Now(),
			time.Now(),
			time.Now(),
		)
	}

	c, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7),
			($8, $9, $10, $11, $12, $13, $14),
			($15, $16, $17, $18, $19, $20, $21)`,
		usersSlice...,
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

	deps := &server.Dependency{
		DB:     db,
		Memory: memory,
	}

	data, err := deps.GetTotal(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data) != "6" {
		t.Errorf("Expected 6, got %s", data)
	}

	// test it again from memory
	data2, err := deps.GetTotal(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data2) != string(data) {
		t.Errorf("Expected %s, got %s", data, data2)
	}
}

func TestGetHourly(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// create a dummy hourly type
	hourly := []server.Hourly{
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
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, &sql.TxOptions{})
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

	deps := &server.Dependency{
		DB:     db,
		Memory: memory,
	}

	data, err := deps.GetHourly(ctx)
	if err != nil {
		t.Error(err)
	}

	data2, err := deps.GetHourly(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(data2) {
		t.Errorf("Expected %s, got %s", data, data2)
	}
}

func TestLastUpdated(t *testing.T) {
	t.Cleanup(Cleanup)

	now := time.Now().Format(time.RFC3339)

	err := memory.Set("analytics:last_updated:users", []byte(now))
	if err != nil {
		t.Error(err)
	}

	err = memory.Set("analytics:last_updated:hourly", []byte(now))
	if err != nil {
		t.Error(err)
	}

	err = memory.Set("analytics:last_updated:total", []byte(now))
	if err != nil {
		t.Error(err)
	}

	deps := &server.Dependency{
		DB:     db,
		Memory: memory,
	}

	data, err := deps.LastUpdated(server.UserEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data.Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data)
	}

	data2, err := deps.LastUpdated(server.HourlyEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data2.Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data2)
	}

	data3, err := deps.LastUpdated(server.TotalEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data3.Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data3)
	}

	// should return an error
	_, err = deps.LastUpdated(server.Endpoint(5))
	if err == nil || !errors.Is(err, errors.New("invalid r value")) {
		t.Error("should error, got none")
	}
}
