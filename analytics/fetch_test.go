package analytics_test

import (
	"context"
	"testing"
	"time"
)

func TestGetUserDataFromDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	data, err := dependency.GetUserDataFromDB(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	data, err := dependency.GetHourlyDataFromDB(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 3 {
		t.Error("data length should be 3, got", len(data))
	}
}
