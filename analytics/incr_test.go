package analytics_test

import (
	"context"
	"teknologi-umum-bot/analytics"
	"testing"
	"time"
)

func TestIncrementUsrDB(t *testing.T) {
	defer Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	users := []analytics.UserMap{
		{
			UserID:      1,
			Username:    "reinaldy",
			DisplayName: "Reinaldy",
			Counter:     10,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
		{
			UserID:      2,
			Username:    "elianiva",
			DisplayName: "Dicha",
			Counter:     20,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
		{
			UserID:      3,
			Username:    "farhan443",
			DisplayName: "Farhan",
			Counter:     15,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
	}

	d := &analytics.Dependency{
		DB:     db,
		Redis:  cache,
		Memory: memory,
	}

	err := d.IncrementUsrDB(ctx, users)
	if err != nil {
		t.Error(err)
	}
}

func TestIncrementUsrRedis(t *testing.T) {
	defer Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	users := analytics.UserMap{
		UserID:      2,
		Username:    "elianiva",
		DisplayName: "Dicha",
		Counter:     20,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		JoinedAt:    time.Now(),
	}

	d := &analytics.Dependency{
		DB:     db,
		Redis:  cache,
		Memory: memory,
	}

	err := d.IncrementUsrRedis(ctx, users)
	if err != nil {
		t.Error(err)
	}
}
