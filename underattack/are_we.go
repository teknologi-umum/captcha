package underattack

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
)

func (d *Dependency) AreWe(ctx context.Context, chatID int64) (bool, error) {
	underAttackCache, err := d.Memory.Get("underattack:" + strconv.FormatInt(chatID, 10))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, err
	}

	if err == nil {
		var entry underattack
		err := json.Unmarshal(underAttackCache, &entry)
		if err != nil {
			return false, err
		}

		return entry.IsUnderAttack && entry.ExpiresAt.After(time.Now()), nil
	}

	underAttackEntry, err := d.GetUnderAttackEntry(ctx, chatID)
	if err != nil {
		return false, err
	}

	marshaledEntry, err := json.Marshal(underAttackEntry)
	if err != nil {
		return false, err
	}

	err = d.Memory.Set("underattack:"+strconv.FormatInt(chatID, 10), marshaledEntry)
	if err != nil {
		return false, err
	}

	return underAttackEntry.IsUnderAttack && underAttackEntry.ExpiresAt.After(time.Now()), nil
}
