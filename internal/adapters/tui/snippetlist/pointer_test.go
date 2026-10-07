package snippetlist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestList_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the clicked row", func(t *testing.T) {
		t.Parallel()

		list, outcomes, _ := listOf(t, numberedSnippets(t, 3)).Update(clickAt(2))

		assert.Equal(t, "Snippet 3", selectedTitle(t, list))
		assert.Empty(t, outcomes)
	})

	t.Run("does nothing below the last row", func(t *testing.T) {
		t.Parallel()

		list, _, _ := listOf(t, numberedSnippets(t, 2)).Update(clickAt(2))

		assert.Equal(t, "Snippet 1", selectedTitle(t, list))
	})
}

func TestList_HasRowAt(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedSnippets(t, 2))

	assert.True(t, list.HasRowAt(pointer.Point{X: 0, Y: 1}))
	assert.False(t, list.HasRowAt(pointer.Point{X: 0, Y: 2}))
}

func TestList_UpdateWheel(t *testing.T) {
	t.Parallel()

	list, _, _ := listOf(t, numberedSnippets(t, 5)).Update(pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 3})

	assert.Contains(t, rows(list)[0], "Snippet 3")
	assert.Equal(t, "Snippet 1", selectedTitle(t, list))
}

func clickAt(y int) pointer.Clicked {
	return pointer.Clicked{At: pointer.Point{X: 0, Y: y}, Double: false}
}
