package captcha

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// Collect AdditionalMsg that was sent because the user did something
// and put it on cache.
//
// It is not recommended using it with a goroutine.
// This should be a normal blocking function.
func (d *Dependencies) collectAdditionalAndCache(captcha *Captcha, m *tb.Message, wrongMsg *tb.Message) error {
	// Because the wrongMsg is another message sent by us, which correlates to the
	// captcha message, we need to put the message ID into the cache.
	// So that we can delete it later.
	captcha.AdditionalMessages = append(captcha.AdditionalMessages, strconv.Itoa(wrongMsg.ID))

	// Update the cache with the added AdditionalMessages
	data, err := json.Marshal(captcha)
	if err != nil {
		return fmt.Errorf("failed to marshal captcha: %w", err)
	}

	err = d.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.FormatInt(m.Chat.ID, 10)+":"+strconv.FormatInt(m.Sender.ID, 10)), data)
		if err != nil {
			txn.Discard()
			return err
		}

		return txn.Commit()
	})
	if err != nil {
		return fmt.Errorf("failed to set captcha in database: %w", err)
	}

	return nil
}

func (d *Dependencies) collectUserMessageAndCache(captcha *Captcha, m *tb.Message) error {
	// We store directly the message ID that was sent by the user into the UserMessages slices.
	captcha.UserMessages = append(captcha.UserMessages, strconv.Itoa(m.ID))

	// Update the cache with the added UserMessages
	data, err := json.Marshal(captcha)
	if err != nil {
		return fmt.Errorf("failed to marshal captcha: %w", err)
	}

	err = d.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.FormatInt(m.Chat.ID, 10)+":"+strconv.FormatInt(m.Sender.ID, 10)), data)
		if err != nil {
			txn.Discard()
			return err
		}

		return txn.Commit()
	})
	if err != nil {
		return fmt.Errorf("failed to set captcha in cache: %w", err)
	}

	return nil
}
