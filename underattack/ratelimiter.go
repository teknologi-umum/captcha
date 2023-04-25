package underattack

import (
	"sync"
	"time"
)

var handlerRateLimiter sync.Map

// RateLimitCall provides a simple rate limiter by a specified ID.
func RateLimitCall(id int64, duration time.Duration) bool {
	_, ok := handlerRateLimiter.Load(id)
	if !ok {
		// Does not exist, we generate one.
		handlerRateLimiter.Store(id, time.Now().Add(duration).Format(time.RFC3339))

		go func(id int64, duration time.Duration) {
			time.Sleep(duration)

			handlerRateLimiter.Delete(id)
		}(id, duration)
		return true
	}

	// Exists, we rate limit them
	return false
}
