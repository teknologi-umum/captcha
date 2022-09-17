package underattack_test

import (
	"context"
	"testing"
	"time"
)

func TestAreWe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	attacked, err := dependency.AreWe(ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !attacked {
		t.Error("we should be attacked, got false")
	}

	cachedAttacked, err := dependency.AreWe(ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cachedAttacked != attacked {
		t.Error("we should be attacked, got false")
	}
}
