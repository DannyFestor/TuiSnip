package snippetlist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

func TestList_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the clicked row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedSnippets(t, 3))

		list, outcomes, _ := list.Update(clickOn(t, list, "Snippet 3"))

		assert.Equal(t, "Snippet 3", selectedTitle(t, list))
		assert.Empty(t, outcomes)
	})

	t.Run("does nothing below the last row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedSnippets(t, 2))
		below := clickOn(t, list, "Snippet 2")
		below.At.Y++

		list, _, _ = list.Update(below)

		assert.Equal(t, "Snippet 1", selectedTitle(t, list))
	})
}

func TestList_HasRowAt(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedSnippets(t, 2))
	last := cellOf(t, list, "Snippet 2")
	below := pointer.Point{X: last.X, Y: last.Y + 1}

	assert.True(t, list.HasRowAt(last))
	assert.False(t, list.HasRowAt(below))
}

func TestList_UpdateWheel(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedSnippets(t, 5))

	list, _, _ = list.Update(pointer.Wheeled{At: cellOf(t, list, "Snippet 2"), Lines: 3})

	assert.Contains(t, rows(list)[0], "Snippet 3")
	assert.Equal(t, "Snippet 1", selectedTitle(t, list))
}

func cellOf(t *testing.T, list snippetlist.List, text string) pointer.Point {
	t.Helper()

	return screencell.Find(t, list.View(upperCursor()), text)
}

func clickOn(t *testing.T, list snippetlist.List, text string) pointer.Clicked {
	t.Helper()

	return pointer.Clicked{At: cellOf(t, list, text), Double: false}
}
