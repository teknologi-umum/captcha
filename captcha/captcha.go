package captcha

import (
	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/badger/v4"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// Dependencies contains the dependency injection struct for
// methods in the captcha package.
type Dependencies struct {
	DB            *badger.DB
	Memory        *bigcache.BigCache
	Bot           *tb.Bot
	TeknumGroupID int64
}
