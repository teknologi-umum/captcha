package underattack_test

import (
	"context"
	"testing"
	"time"
)

func TestGetUnderAttackEntry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	entry, err := dependency.GetUnderAttackEntry(ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if entry.IsUnderAttack == false {
		t.Error("expecting IsUnderAttack to be true, got false")
	}

	if entry.ExpiresAt.Before(time.Now()) {
		t.Errorf("expecting ExpiresAt to be after now, got: %v", entry.ExpiresAt)
	}

	if entry.NotificationMessageID != 1002 {
		t.Errorf("expecting NotificationMessageID to be 1002, got: %v", entry.NotificationMessageID)
	}
}

func TestGetUnderAttackEntry_NotExists(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := dependency.GetUnderAttackEntry(ctx, 20)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateNewEntry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := dependency.CreateNewEntry(ctx, 2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetUnderAttackStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := dependency.SetUnderAttackStatus(ctx, 3, true, time.Now().Add(time.Minute*30), 1003)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
