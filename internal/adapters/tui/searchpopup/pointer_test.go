package searchpopup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestPopup_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("reveals the clicked result and closes", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Click("Prune everything")

		assert.Equal(t, []outcome.Outcome{outcome.SnippetRevealed{ID: snippets[1].ID()}}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("does nothing on a click in the preview", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Click("Stop accepting")

		assert.Empty(t, screen.Outcomes())
		assert.True(t, screen.IsOpen())
	})

	t.Run("closes without an outcome on a click outside", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})
}
