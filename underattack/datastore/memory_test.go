package datastore_test

import (
	"context"
	"encoding/json"
	"github.com/teknologi-umum/captcha/underattack"
	"github.com/teknologi-umum/captcha/underattack/datastore"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
)

func TestNewInMemoryDatastore(t *testing.T) {
	var dependency underattack.Datastore

	db, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour))
	if err != nil {
		t.Fatalf("Creating bigcache instance: %s", err.Error())
	}

	dependency, err = datastore.NewInMemoryDatastore(db)
	if err != nil {
		t.Fatalf("creating new postgres datastore: %s", err.Error())
	}

	setupCtx, setupCancel := context.WithTimeout(context.Background(), time.Second*30)

	err = dependency.Migrate(setupCtx)
	if err != nil {
		t.Fatalf("migrating tables: %s", err.Error())
	}

	err = SeedMemoryDatastore(setupCtx, db)
	if err != nil {
		t.Fatalf("seeding data: %s", err.Error())
	}

	t.Cleanup(func() {
		setupCancel()

		err := dependency.Close()
		if err != nil {
			t.Logf("closing postgres database: %s", err.Error())
		}
	})

	t.Run("NewInMemoryDatastore", func(t *testing.T) {
		t.Run("Nil DB", func(t *testing.T) {
			_, err := datastore.NewInMemoryDatastore(nil)
			if err.Error() != "nil db" {
				t.Errorf("expecting an error of 'nil db', instead got %s", err.Error())
			}
		})
	})

	t.Run("Migrate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := dependency.Migrate(ctx)
		if err != nil {
			t.Errorf("migrating database: %s", err.Error())
		}
	})

	t.Run("GetUnderAttackEntry", func(t *testing.T) {
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
	})

	t.Run("GetUnderAttackEntry_NotExists", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		_, err := dependency.GetUnderAttackEntry(ctx, 20)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("CreateNewEntry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := dependency.CreateNewEntry(ctx, 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SetUnderAttackStatus", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := dependency.SetUnderAttackStatus(ctx, 3, true, time.Now().Add(time.Minute*30), 1003)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func SeedMemoryDatastore(ctx context.Context, db *bigcache.BigCache) error {
	value, err := json.Marshal(underattack.UnderAttack{
		GroupID:               1,
		IsUnderAttack:         true,
		NotificationMessageID: 1002,
		ExpiresAt:             time.Now().Add(time.Hour),
		UpdatedAt:             time.Now(),
	})
	if err != nil {
		return err
	}

	return db.Set("1", value)
}
