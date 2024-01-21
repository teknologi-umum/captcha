package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// Cleanup will iterate over every keys and make sure the expiry has not been exceeded by 10 seconds.
// If it is, we'll kick the person.
func (d *Dependencies) Cleanup() {
	var captchaPrefix = []byte("captcha:")
	ctx := context.Background()

	for {
		var captchas []Captcha
		err := d.DB.View(func(txn *badger.Txn) error {
			iteratorOptions := badger.DefaultIteratorOptions
			iteratorOptions.PrefetchSize = 10
			iterator := txn.NewIterator(iteratorOptions)
			defer iterator.Close()
			for iterator.Rewind(); iterator.Valid(); iterator.Next() {
				item := iterator.Item()
				key := item.Key()
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}

				if bytes.HasPrefix(key, captchaPrefix) {
					// We don't need these
					continue
				}

				// Try to read the value as captcha struct
				var captcha Captcha
				err = json.Unmarshal(value, &captcha)
				if err != nil {
					// What's the point of reporting if it can't be parsed?
					continue
				}

				captchas = append(captchas, captcha)
			}

			return nil
		})
		if err != nil {
			sentry.CaptureException(err)
			time.Sleep(time.Minute * 5)
			continue
		}

		for _, captcha := range captchas {
			err := d.removeUserFromGroup(ctx, &tb.Chat{ID: captcha.ChatID}, &tb.User{ID: captcha.SenderID}, captcha)
			if err != nil {
				sentry.CaptureException(err)
			}
		}

		time.Sleep(time.Minute * 5)
		continue
	}
}
