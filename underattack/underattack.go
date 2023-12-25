package underattack

import (
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	tb "gopkg.in/telebot.v3"
)

// Dependency contains the dependency injection struct
// for methods in the underattack package
type Dependency struct {
	Memory *bigcache.BigCache
	DB     *sqlx.DB
	Bot    *tb.Bot
}

// underattack provides a data struct to interact with
// the database table.
type underattack struct {
	GroupID               int64     `db:"group_id"`
	IsUnderAttack         bool      `db:"is_under_attack"`
	NotificationMessageID int64     `db:"notification_message_id"`
	ExpiresAt             time.Time `db:"expires_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}
