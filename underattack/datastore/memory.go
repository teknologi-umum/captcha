package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/underattack"

	"github.com/allegro/bigcache/v3"
)

type memoryDatastore struct {
	db *bigcache.BigCache
}

func NewInMemoryDatastore(db *bigcache.BigCache) (underattack.Datastore, error) {
	if db == nil {
		return nil, fmt.Errorf("nil db")
	}

	return &memoryDatastore{db: db}, nil
}

func (m *memoryDatastore) Migrate(ctx context.Context) error {
	// Nothing to migrate
	return nil
}

func (m *memoryDatastore) GetUnderAttackEntry(ctx context.Context, groupID int64) (underattack.UnderAttack, error) {
	span := sentry.StartSpan(ctx, "memory_datastore.get_under_attack_entry")
	defer span.Finish()
	ctx = span.Context()

	value, err := m.db.Get(strconv.FormatInt(groupID, 10))
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			go func(groupID int64) {
				ctx := sentry.SetHubOnContext(context.Background(), sentry.GetHubFromContext(ctx))

				time.Sleep(time.Second * 5)
				ctx, cancel := context.WithTimeout(ctx, time.Second*15)
				defer cancel()

				err := m.CreateNewEntry(ctx, groupID)
				if err != nil {
					sentry.GetHubFromContext(ctx).CaptureException(err)
				}
			}(groupID)

			return underattack.UnderAttack{}, nil
		}

		return underattack.UnderAttack{}, err
	}

	var entry underattack.UnderAttack
	err = json.Unmarshal(value, &entry)
	if err != nil {
		return underattack.UnderAttack{}, err
	}

	return entry, nil
}

func (m *memoryDatastore) CreateNewEntry(ctx context.Context, groupID int64) error {
	span := sentry.StartSpan(ctx, "memory_datastore.create_new_entry")
	defer span.Finish()

	if _, err := m.db.Get(strconv.FormatInt(groupID, 10)); err != nil {
		// Do nothing if already exists
		return nil
	}

	// Set a new one if not exists
	value, err := json.Marshal(underattack.UnderAttack{
		GroupID:               groupID,
		IsUnderAttack:         false,
		NotificationMessageID: 0,
		ExpiresAt:             time.Time{},
		UpdatedAt:             time.Now(),
	})
	if err != nil {
		return err
	}

	return m.db.Set(strconv.FormatInt(groupID, 10), value)
}

func (m *memoryDatastore) SetUnderAttackStatus(ctx context.Context, groupID int64, underAttack bool, expiresAt time.Time, notificationMessageID int64) error {
	span := sentry.StartSpan(ctx, "memory_datastore.set_under_attack_status")
	defer span.Finish()

	// Set a new one if not exists
	value, err := json.Marshal(underattack.UnderAttack{
		GroupID:               groupID,
		IsUnderAttack:         underAttack,
		NotificationMessageID: notificationMessageID,
		ExpiresAt:             expiresAt,
		UpdatedAt:             time.Now(),
	})
	if err != nil {
		return err
	}

	return m.db.Set(strconv.FormatInt(groupID, 10), value)
}

func (m *memoryDatastore) Close() error {
	return m.db.Close()
}
