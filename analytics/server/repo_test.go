package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/teknologi-umum/captcha/analytics/server"
)

func TestGetAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	res, err := dependency.GetAll(ctx)
	if err != nil {
		t.Error(err)
	}

	var user []server.User
	err = json.Unmarshal(res, &user)
	if err != nil {
		t.Error(err)
	}

	if len(user) != 3 {
		t.Fatalf("Expected 3 user, got %d", len(user))
	}

	if user[0].UserID != 500 {
		t.Error("user id should be 500, got:", user[0].UserID)
	}

	if user[0].Username != "user5" {
		t.Error("username should be user5, got:", user[0].Username)
	}

	if user[0].DisplayName != "User 5" {
		t.Error("display name should be User 5, got:", user[0].DisplayName)
	}

	if user[0].Counter != 1 {
		t.Error("counter should be 1, got:", user[0].Counter)
	}

	// try to get it again
	res2, err := dependency.GetAll(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(res) != string(res2) {
		t.Errorf("result should be the same\n\nres1: %s\n\nres2: %s", string(res), string(res2))
	}
}

func TestGetTotal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	data, err := dependency.GetTotal(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data) != "6" {
		t.Errorf("Expected 6, got %s", data)
	}

	// test it again from memory
	data2, err := dependency.GetTotal(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data2) != string(data) {
		t.Errorf("Expected %s, got %s", data, data2)
	}
}

func TestGetHourly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	data, err := dependency.GetHourly(ctx)
	if err != nil {
		t.Error(err)
	}

	data2, err := dependency.GetHourly(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(data2) {
		t.Errorf("Expected %s, got %s", data, data2)
	}
}

func TestGetDukunPoints(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := dependency.GetDukunPoints(ctx)
	if err != nil {
		t.Error(err)
	}

	data2, err := dependency.GetDukunPoints(ctx)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(data2) {
		t.Errorf("Expected %s, got %s", data, data2)
	}
}

func TestLastUpdated(t *testing.T) {
	now := time.Now().Round(time.Minute).Format(time.RFC3339)

	data, err := dependency.LastUpdated(server.UserEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data.Round(time.Minute).Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data)
	}

	data2, err := dependency.LastUpdated(server.HourlyEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data2.Round(time.Minute).Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data2)
	}

	data3, err := dependency.LastUpdated(server.TotalEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data3.Round(time.Minute).Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data3)
	}

	data4, err := dependency.LastUpdated(server.DukunEndpoint)
	if err != nil {
		t.Error(err)
	}

	if data4.Round(time.Minute).Format(time.RFC3339) != now {
		t.Errorf("Expected %s, got %s", now, data4)
	}

	// should return an error
	_, err = dependency.LastUpdated(server.Endpoint(5))
	if err == nil {
		t.Error("should error, got none")
	}

	if !errors.Is(err, server.ErrInvalidValue) {
		t.Error("should error with ErrInvalidValue, got:", err)
	}
}
