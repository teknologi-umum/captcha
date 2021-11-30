package captcha

import (
	"errors"
	"strings"

	"github.com/allegro/bigcache/v3"
)

// Check if a cache with a specific key exists or not.
func cacheExists(cache *bigcache.BigCache, key string) bool {
	_, err := cache.Get(key)
	return !errors.Is(err, bigcache.ErrEntryNotFound)
}

// Check if a user exists on the "captcha:users" key.
func userExists(cache *bigcache.BigCache, key string) (bool, error) {
	users, err := cache.Get("captcha:users")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, err
	}

	// Split the users which is in the type of []byte
	// to []string first. Then we'll iterate through it.
	// Also, we'd like to pop the last array, because it's
	// just an empty string.
	str := strings.Split(string(users), ";")
	for _, v := range str {
		if v == key {
			return true, nil
		}
	}
	return false, nil
}
