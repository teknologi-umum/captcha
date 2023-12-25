package underattack

import (
	"time"

	"github.com/allegro/bigcache/v3"
	tb "gopkg.in/telebot.v3"
)

// Dependency contains the dependency injection struct
// for methods in the UnderAttack package
type Dependency struct {
	Datastore Datastore
	Memory    *bigcache.BigCache
	Bot       *tb.Bot
}

// UnderAttack provides a data struct to interact with
// the database table.
type UnderAttack struct {
	GroupID               int64     `db:"group_id"`
	IsUnderAttack         bool      `db:"is_under_attack"`
	NotificationMessageID int64     `db:"notification_message_id"`
	ExpiresAt             time.Time `db:"expires_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}
