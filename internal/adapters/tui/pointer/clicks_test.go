package pointer_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

const column = 3

func TestClicks_Next(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)

	t.Run("reports a first click as single", func(t *testing.T) {
		t.Parallel()

		_, clicked, ok := pointer.Clicks{}.Next(leftClickOnLine(4), start)

		assert.True(t, ok)
		assert.Equal(t, pointer.Clicked{At: pointer.Point{X: column, Y: 4}, Double: false}, clicked)
	})

	t.Run("reports a second click on the same cell soon after as double", func(t *testing.T) {
		t.Parallel()

		clicks, _, _ := pointer.Clicks{}.Next(leftClickOnLine(4), start)
		_, clicked, _ := clicks.Next(leftClickOnLine(4), start.Add(300*time.Millisecond))

		assert.True(t, clicked.Double)
	})

	t.Run("reports a second click on another cell as single", func(t *testing.T) {
		t.Parallel()

		clicks, _, _ := pointer.Clicks{}.Next(leftClickOnLine(4), start)
		_, clicked, _ := clicks.Next(leftClickOnLine(5), start.Add(300*time.Millisecond))

		assert.False(t, clicked.Double)
	})

	t.Run("reports a second click after the pause as single", func(t *testing.T) {
		t.Parallel()

		clicks, _, _ := pointer.Clicks{}.Next(leftClickOnLine(4), start)
		_, clicked, _ := clicks.Next(leftClickOnLine(4), start.Add(time.Second))

		assert.False(t, clicked.Double)
	})

	t.Run("starts over after a double click", func(t *testing.T) {
		t.Parallel()

		clicks, _, _ := pointer.Clicks{}.Next(leftClickOnLine(4), start)
		clicks, _, _ = clicks.Next(leftClickOnLine(4), start.Add(100*time.Millisecond))
		_, clicked, _ := clicks.Next(leftClickOnLine(4), start.Add(200*time.Millisecond))

		assert.False(t, clicked.Double)
	})

	t.Run("ignores every button but the left one", func(t *testing.T) {
		t.Parallel()

		_, _, ok := pointer.Clicks{}.Next(tea.MouseClickMsg{X: column, Y: 4, Button: tea.MouseRight, Mod: 0}, start)

		assert.False(t, ok)
	})
}

func leftClickOnLine(line int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: column, Y: line, Button: tea.MouseLeft, Mod: 0}
}
