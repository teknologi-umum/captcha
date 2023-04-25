package underattack_test

import (
	"testing"
	"time"

	"teknologi-umum-bot/underattack"
)

func TestRateLimitCall(t *testing.T) {
	id := int64(123)
	duration := time.Second

	rateLimited := underattack.RateLimitCall(id, duration)
	if !rateLimited {
		t.Error("expecting true, got false")
	}

	rateLimited = underattack.RateLimitCall(id, duration)
	if rateLimited {
		t.Error("expecting false, got true")
	}

}
