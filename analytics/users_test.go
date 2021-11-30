package analytics_test

import (
	"context"
	"teknologi-umum-bot/analytics"
	"testing"
	"time"
)

func TestGetAllUserID(t *testing.T) {
	defer Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	err := cache.SAdd(ctx, "analytics:users", "Adam", "Bobby", "Clifford").Err()
	if err != nil {
		t.Error(err)
	}

	deps := &analytics.Dependency{
		DB:     db,
		Redis:  cache,
		Memory: memory,
	}

	users, err := deps.GetAllUserID(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(users) != 3 {
		t.Error("Expected 3 users, got ", len(users))
	}
}

func TestGetAllUserMap(t *testing.T) {
	defer Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	tx := cache.TxPipeline()
	defer tx.Close()
	tx.SAdd(ctx, "analytics:users", "1", "2", "3")
	tx.HSet(ctx, "analytics:1", "username", "adam", "display_name", "Adam", "counter", 1)
	tx.HSet(ctx, "analytics:2", "username", "bobby45", "display_name", "Bobby", "counter", 5)
	tx.HSet(ctx, "analytics:3", "username", "clifford77", "display_name", "Clifford", "counter", 3)

	_, err := tx.Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	deps := &analytics.Dependency{
		DB:     db,
		Redis:  cache,
		Memory: memory,
	}

	users, err := deps.GetAllUserMap(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(users) != 3 {
		t.Error("Expected 3 users, got ", len(users))
	}
}

func TestFlushAllUserID(t *testing.T) {
	defer Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	tx := cache.TxPipeline()
	defer tx.Close()
	tx.SAdd(ctx, "analytics:users", "1", "2", "3")
	tx.HSet(ctx, "analytics:1", "username", "adam", "display_name", "Adam", "counter", 1)
	tx.HSet(ctx, "analytics:2", "username", "bobby45", "display_name", "Bobby", "counter", 5)
	tx.HSet(ctx, "analytics:3", "username", "clifford77", "display_name", "Clifford", "counter", 3)

	_, err := tx.Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	deps := &analytics.Dependency{
		DB:     db,
		Redis:  cache,
		Memory: memory,
	}

	err = deps.FlushAllUserID(ctx)
	if err != nil {
		t.Error(err)
	}
}
