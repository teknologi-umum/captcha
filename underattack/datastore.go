package underattack

import (
	"context"
	"time"
)

type Datastore interface {
	Migrate(ctx context.Context) error
	GetUnderAttackEntry(ctx context.Context, groupID int64) (UnderAttack, error)
	CreateNewEntry(ctx context.Context, groupID int64) error
	SetUnderAttackStatus(ctx context.Context, groupID int64, underAttack bool, expiresAt time.Time, notificationMessageID int64) error
	Close() error
}
