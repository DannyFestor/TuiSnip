package testkit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestManualClock_Advance(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	clock := testkit.NewManualClock(start)

	clock.Advance(time.Minute)

	assert.Equal(t, start.Add(time.Minute), clock.Now())
}

func TestFixedClock_Now(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	assert.Equal(t, now, testkit.NewFixedClock(now).Now())
}
