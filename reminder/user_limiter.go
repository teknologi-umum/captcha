package reminder

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
)

func (d *Dependency) CheckUserLimit(ctx context.Context, id int64) (n int, err error) {
	span := sentry.StartSpan(ctx, "reminder.check_user_limiter")
	defer span.Finish()

	value, err := d.memory.Get(fmt.Sprintf("reminder:user_limit:%d", id))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return 0, fmt.Errorf("acquiring value from memory: %w", err)
	}

	if value == nil || string(value) == "" {
		value = []byte("0")
	}

	return strconv.Atoi(string(value))
}

func (d *Dependency) IncrementUserLimit(ctx context.Context, id int64) error {
	span := sentry.StartSpan(ctx, "reminder.increment_user_limit")
	defer span.Finish()

	value, err := d.memory.Get(fmt.Sprintf("reminder:user_limit:%d", id))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return fmt.Errorf("acquiring value from memory: %w", err)
	}

	if value == nil || string(value) == "" {
		value = []byte("0")
	}

	i, err := strconv.Atoi(string(value))
	if err != nil {
		return fmt.Errorf("invalid value: %s", value)
	}

	err = d.memory.Set(fmt.Sprintf("reminder:user_limit:%d", id), []byte(strconv.Itoa(i+1)))
	if err != nil {
		return fmt.Errorf("setting value to memory: %w", err)
	}

	return nil
}

func (d *Dependency) DecrementUserLimit(ctx context.Context, id int64) error {
	span := sentry.StartSpan(ctx, "reminder.decrement_user_limit")
	defer span.Finish()

	value, err := d.memory.Get(fmt.Sprintf("reminder:user_limit:%d", id))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return fmt.Errorf("acquiring value from memory: %w", err)
	}

	if value == nil || string(value) == "" {
		// Don't decrement anything that's empty
		return nil
	}

	i, err := strconv.Atoi(string(value))
	if err != nil {
		return fmt.Errorf("invalid value: %s", value)
	}

	err = d.memory.Set(fmt.Sprintf("reminder:user_limit:%d", id), []byte(strconv.Itoa(i-1)))
	if err != nil {
		return fmt.Errorf("setting value to memory: %w", err)
	}

	return nil
}
