package tagpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

func TestPane_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the clicked row", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane := samplePane(t, sample)

		pane, outcomes, _ := pane.Update(clickOn(t, pane, "# go"))

		selected, _ := pane.Selected()
		assert.Equal(t, sample.golang.ID(), selected)
		assert.Empty(t, outcomes)
	})

	t.Run("does nothing below the last row", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane := paneIn(t, look.Size{Width: boxWidth, Height: 5}).WithTags(sample.counts)
		below := clickOn(t, pane, "# unused")
		below.At.Y++

		pane, _, _ = pane.Update(below)

		selected, _ := pane.Selected()
		assert.Equal(t, sample.docker.ID(), selected)
	})

	t.Run("ignores a click while a name is typed", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(samplePane(t, newSampleTags(t)), newTagTyped("ci")...)

		pane, _, _ = pane.Update(clickOn(t, pane, "# go"))

		assert.True(t, pane.Naming())
	})
}

func TestPane_HasRowAt(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: 5}).WithTags(newSampleTags(t).counts)
	last := cellOf(t, pane, "# unused")
	below := pointer.Point{X: last.X, Y: last.Y + 1}

	assert.True(t, pane.HasRowAt(last))
	assert.False(t, pane.HasRowAt(below))
}

func TestPane_UpdateWheel(t *testing.T) {
	t.Parallel()

	sample := newSampleTags(t)
	pane := paneIn(t, look.Size{Width: boxWidth, Height: 2}).WithTags(sample.counts)

	pane, _, _ = pane.Update(pointer.Wheeled{At: cellOf(t, pane, "# go"), Lines: 3})

	selected, _ := pane.Selected()
	assert.Equal(t, []string{"# go           4", "# unused       0"}, viewLines(pane))
	assert.Equal(t, sample.docker.ID(), selected)
}

func cellOf(t *testing.T, pane tagpane.Pane, text string) pointer.Point {
	t.Helper()

	return screencell.Find(t, pane.View(upperCursor()), text)
}

func clickOn(t *testing.T, pane tagpane.Pane, text string) pointer.Clicked {
	t.Helper()

	return pointer.Clicked{At: cellOf(t, pane, text), Double: false}
}
