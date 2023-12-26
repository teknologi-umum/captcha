package reminder_test

import (
	"os"
	"testing"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/reminder"
)

func TestMain(m *testing.M) {
	_ = sentry.Init(sentry.ClientOptions{})

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	t.Run("Nil memory", func(t *testing.T) {
		_, err := reminder.New(nil)
		if err == nil {
			t.Errorf("expecting an error, got nil")
		} else {
			if err.Error() != "memory is nil" {
				t.Errorf("expecting an error of 'memory is nil', instead got %v", err)
			}
		}
	})

	t.Run("Happy", func(t *testing.T) {
		dependency, err := reminder.New(&bigcache.BigCache{})
		if err != nil {
			t.Errorf("expecting a nil error, instead got %v", err)
		} else {
			if dependency == nil {
				t.Errorf("exepcting dependency to not be nil, got nil instead")
			}
		}
	})
}
