package confirm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestConfirmation_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("answers no on a click outside", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("stays open on a click inside", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		screen.Click(question)

		assert.True(t, screen.IsOpen())
	})
}
