package captcha

import (
	"errors"
	"strconv"
	"strings"

	"github.com/allegro/bigcache/v3"
)

// Check if a cache with a specific key exists or not.
func (d *Dependencies) cacheExists(key string) bool {
	_, err := d.Memory.Get(key)
	return !errors.Is(err, bigcache.ErrEntryNotFound)
}

// Check if a user exists on the "captcha:users" key.
func (d *Dependencies) userExists(userID int64, groupID int64) (bool, error) {
	users, err := d.Memory.Get("captcha:users:" + strconv.FormatInt(groupID, 10))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, err
	}

	// Split the users which is in the type of []byte
	// to []string first. Then we'll iterate through it.
	// Also, we'd like to pop the last array, because it's
	// just an empty string.
	str := strings.Split(string(users), ";")
	key := strconv.FormatInt(userID, 10)
	for _, v := range str {
		if v == key {
			return true, nil
		}
	}
	return false, nil
}
