package analytics_test

import (
	"context"
	"teknologi-umum-bot/analytics"
	"testing"
	"time"
)

func TestIncrementUsrDB(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	users := []analytics.GroupMember{
		{
			UserID:      1,
			GroupID:     analytics.NullInt64{Int64: 5, Valid: true},
			Username:    "reinaldy",
			DisplayName: "Reinaldy",
			Counter:     10,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
		{
			UserID:      2,
			GroupID:     analytics.NullInt64{Int64: 5, Valid: true},
			Username:    "elianiva",
			DisplayName: "Dicha",
			Counter:     20,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
		{
			UserID:      3,
			GroupID:     analytics.NullInt64{Int64: 5, Valid: true},
			Username:    "farhan443",
			DisplayName: "Farhan",
			Counter:     15,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			JoinedAt:    time.Now(),
		},
	}

	d := &analytics.Dependency{
		DB:       db,
		Memory:   memory,
		TeknumID: "123456789",
	}

	err := d.IncrementUserDB(ctx, users[0])
	if err != nil {
		t.Error(err)
	}
}
