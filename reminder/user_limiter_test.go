package reminder_test

import (
	"context"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/reminder"
)

func TestUserLimit(t *testing.T) {
	memory, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		t.Fatalf("Creating cache: %v", err)
	}
	t.Cleanup(func() {
		_ = memory.Close()
	})

	dependency, err := reminder.New(memory)
	if err != nil {
		t.Fatalf("Creating reminder instance: %s", err.Error())
	}

	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub())

	userLimit, err := dependency.CheckUserLimit(ctx, 123)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if userLimit != 0 {
		t.Errorf("expecting userLimit to be 0, instead got %d", userLimit)
	}

	err = dependency.IncrementUserLimit(ctx, 123, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	userLimit2, err := dependency.CheckUserLimit(ctx, 123)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	if userLimit2 != 1 {
		t.Errorf("expecting userLimit2 to be 1, instead got %d", userLimit2)
	}
}
