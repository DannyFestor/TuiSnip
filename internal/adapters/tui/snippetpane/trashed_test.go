package snippetpane_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
)

func TestPane_ShowingInTrash(t *testing.T) {
	t.Parallel()

	trashed := snippetpane.Trashed{
		Place:     "Root / go / testing",
		DeletedAt: time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC),
	}

	t.Run("shows the path it was in and the day it was deleted", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, time.UTC, look.Size{Width: 80, Height: 10}).ShowingInTrash(longSnippet(t), trashed)

		got := trimmed(lines(pane))

		assert.Equal(t, "Root / go / testing · Go", got[1])
		assert.Equal(t, "created 2026-09-06 · updated 2026-09-06 · deleted 2026-10-14", got[2])
	})

	t.Run("asks to Copy from the Trash", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, time.UTC, look.Size{Width: 80, Height: 10}).ShowingInTrash(longSnippet(t), trashed)

		request, ok := pane.Copying()

		require.True(t, ok)
		assert.True(t, request.InTrash)
	})

	t.Run("shows a live Snippet again without the Trash details", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, time.UTC, look.Size{Width: 80, Height: 10}).
			ShowingInTrash(longSnippet(t), trashed).
			Showing(longSnippet(t))

		request, _ := pane.Copying()

		assert.Equal(t, "Root · Go", trimmed(lines(pane))[1])
		assert.False(t, request.InTrash)
	})
}
