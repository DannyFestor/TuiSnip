package system_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
)

func TestClock_Now(t *testing.T) {
	t.Parallel()

	before := time.Now()

	now := system.NewClock().Now()

	assert.WithinRange(t, now, before, time.Now())
}
