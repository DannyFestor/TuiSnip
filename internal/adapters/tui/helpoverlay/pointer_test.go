package helpoverlay_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestOverlay_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("closes on a click outside", func(t *testing.T) {
		t.Parallel()

		screen := helping(t)

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.False(t, screen.IsOpen())
	})

	t.Run("stays open on a click inside", func(t *testing.T) {
		t.Parallel()

		screen := helping(t)

		screen.Click("Help")

		assert.True(t, screen.IsOpen())
	})
}
