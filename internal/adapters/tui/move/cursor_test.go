package move_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
)

const (
	rows   = 10
	height = 3
)

func TestCursor_Moved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		start      int
		direction  move.Direction
		wantIndex  int
		wantOffset int
	}{
		{name: "down moves one row", start: 0, direction: move.Down, wantIndex: 1, wantOffset: 0},
		{name: "down scrolls to keep the cursor in view", start: 2, direction: move.Down, wantIndex: 3, wantOffset: 1},
		{name: "down stops at the last row", start: rows - 1, direction: move.Down, wantIndex: rows - 1, wantOffset: 7},
		{name: "up stops at the first row", start: 0, direction: move.Up, wantIndex: 0, wantOffset: 0},
		{name: "top goes to the first row", start: 5, direction: move.Top, wantIndex: 0, wantOffset: 0},
		{name: "bottom goes to the last row", start: 0, direction: move.Bottom, wantIndex: rows - 1, wantOffset: 7},
		{name: "page down moves a page", start: 0, direction: move.PageDown, wantIndex: height, wantOffset: 1},
		{name: "page up moves a page", start: 5, direction: move.PageUp, wantIndex: 2, wantOffset: 2},
		{name: "none stays", start: 4, direction: move.None, wantIndex: 4, wantOffset: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := move.Cursor{}.At(tt.start, rows, height).Moved(tt.direction, rows, height)

			assert.Equal(t, tt.wantIndex, got.Index())
			assert.Equal(t, tt.wantOffset, got.Offset())
		})
	}
}

func TestCursor_End(t *testing.T) {
	t.Parallel()

	t.Run("ends a page after the scrolled row", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 5, move.Cursor{}.At(4, rows, height).End(rows, height))
	})

	t.Run("stops at the last row", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 2, move.Cursor{}.At(1, 2, height).End(2, height))
	})
}

func TestCursor_At(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		index     int
		rows      int
		wantIndex int
	}{
		{name: "keeps an index inside the rows", index: 4, rows: rows, wantIndex: 4},
		{name: "clamps past the last row", index: rows + 5, rows: rows, wantIndex: rows - 1},
		{name: "clamps below the first row", index: -2, rows: rows, wantIndex: 0},
		{name: "rests on the first row without rows", index: 3, rows: 0, wantIndex: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantIndex, move.Cursor{}.At(tt.index, tt.rows, height).Index())
		})
	}
}
