package tagpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestPane_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the clicked row", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)

		pane, outcomes, _ := samplePane(t, sample).Update(clickAt(1))

		selected, _ := pane.Selected()
		assert.Equal(t, sample.golang.ID(), selected)
		assert.Empty(t, outcomes)
	})

	t.Run("does nothing below the last row", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane := paneIn(t, look.Size{Width: boxWidth, Height: 5}).WithTags(sample.counts)

		pane, _, _ = pane.Update(clickAt(3))

		selected, _ := pane.Selected()
		assert.Equal(t, sample.docker.ID(), selected)
	})

	t.Run("ignores a click while a name is typed", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(samplePane(t, newSampleTags(t)), newTagTyped("ci")...)

		pane, _, _ = pane.Update(clickAt(1))

		assert.True(t, pane.Naming())
	})
}

func TestPane_HasRowAt(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: 5}).WithTags(newSampleTags(t).counts)

	assert.True(t, pane.HasRowAt(pointer.Point{X: 0, Y: 2}))
	assert.False(t, pane.HasRowAt(pointer.Point{X: 0, Y: 3}))
}

func TestPane_UpdateWheel(t *testing.T) {
	t.Parallel()

	sample := newSampleTags(t)
	pane := paneIn(t, look.Size{Width: boxWidth, Height: 2}).WithTags(sample.counts)

	pane, _, _ = pane.Update(pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 3})

	selected, _ := pane.Selected()
	assert.Equal(t, []string{"# go           4", "# unused       0"}, viewLines(pane))
	assert.Equal(t, sample.docker.ID(), selected)
}

func clickAt(y int) pointer.Clicked {
	return pointer.Clicked{At: pointer.Point{X: 0, Y: y}, Double: false}
}
