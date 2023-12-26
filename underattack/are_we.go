package underattack

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/allegro/bigcache/v3"
)

// AreWe ...on under attack mode?
func (d *Dependency) AreWe(ctx context.Context, chatID int64) (bool, error) {
	span := sentry.StartSpan(ctx, "underattack.are_we", sentry.WithTransactionName("Are we under attack?"))
	defer span.Finish()
	ctx = span.Context()

	underAttackCache, err := d.Memory.Get("UnderAttack:" + strconv.FormatInt(chatID, 10))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, err
	}

	if err == nil {
		var entry UnderAttack
		err := json.Unmarshal(underAttackCache, &entry)
		if err != nil {
			return false, err
		}

		return entry.IsUnderAttack && entry.ExpiresAt.After(time.Now()), nil
	}

	underAttackEntry, err := d.Datastore.GetUnderAttackEntry(ctx, chatID)
	if err != nil {
		return false, err
	}

	marshaledEntry, err := json.Marshal(underAttackEntry)
	if err != nil {
		return false, err
	}

	err = d.Memory.Set("UnderAttack:"+strconv.FormatInt(chatID, 10), marshaledEntry)
	if err != nil {
		return false, err
	}

	return underAttackEntry.IsUnderAttack && underAttackEntry.ExpiresAt.After(time.Now()), nil
}
