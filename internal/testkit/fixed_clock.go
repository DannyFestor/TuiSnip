package testkit

import "time"

type FixedClock struct {
	now time.Time
}

func NewFixedClock(now time.Time) FixedClock {
	return FixedClock{now: now}
}

func (c FixedClock) Now() time.Time {
	return c.now
}
