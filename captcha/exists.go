package captcha

import (
	"errors"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Check if a cache with a specific key exists or not.
func (d *Dependencies) cacheExists(key string) bool {
	err := d.DB.View(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(key)); err != nil {
			return err
		}
		return nil
	})
	return !errors.Is(err, badger.ErrKeyNotFound)
}

// Check if a user exists on the "captcha:users" key.
func (d *Dependencies) userExists(userID int64, groupID int64) (exists bool, err error) {
	err = d.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("captcha:users:" + strconv.FormatInt(groupID, 10)))
		if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Split the users which is in the type of []byte
		// to []string first. Then we'll iterate through it.
		// Also, we'd like to pop the last array, because it's
		// just an empty string.
		str := strings.Split(string(value), ";")
		key := strconv.FormatInt(userID, 10)
		for _, v := range str {
			if v == key {
				exists = true
				break
			}
		}

		return nil
	})
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return false, err
	}

	return exists, nil
}
