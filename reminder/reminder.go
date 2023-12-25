package reminder

import (
	"fmt"
	"github.com/allegro/bigcache/v3"
	"time"
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
