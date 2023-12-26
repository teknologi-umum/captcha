package reminder

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

type Dependency struct {
	memory *bigcache.BigCache
}

func New(memory *bigcache.BigCache) (*Dependency, error) {
	if memory == nil {
		return nil, fmt.Errorf("memory is nil")
	}

	return &Dependency{memory: memory}, nil
}

type Reminder struct {
	Subject []string // maximum of 3 person ping
	Time    time.Time
	Object  string
}

var ErrExceeds24Hours = fmt.Errorf("exceeds 24 hours")
