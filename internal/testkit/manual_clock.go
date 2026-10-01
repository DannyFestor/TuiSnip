package testkit

import (
	"sync"
	"time"
)

type ManualClock struct {
	mu  sync.RWMutex
	now time.Time
}

func NewManualClock(start time.Time) *ManualClock {
	return &ManualClock{mu: sync.RWMutex{}, now: start}
}

func (c *ManualClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.now
}

func (c *ManualClock) Advance(step time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(step)
}
