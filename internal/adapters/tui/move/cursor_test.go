package move_test

import (
	"strconv"
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
		name      string
		start     int
		direction move.Direction
		wantShown string
	}{
		{name: "down moves one row", start: 0, direction: move.Down, wantShown: "0\n[1]\n2"},
		{name: "down scrolls to keep the cursor in view", start: 2, direction: move.Down, wantShown: "1\n2\n[3]"},
		{name: "down stops at the last row", start: rows - 1, direction: move.Down, wantShown: "7\n8\n[9]"},
		{name: "up stops at the first row", start: 0, direction: move.Up, wantShown: "[0]\n1\n2"},
		{name: "top goes to the first row", start: 5, direction: move.Top, wantShown: "[0]\n1\n2"},
		{name: "bottom goes to the last row", start: 0, direction: move.Bottom, wantShown: "7\n8\n[9]"},
		{name: "page down moves a page", start: 0, direction: move.PageDown, wantShown: "1\n2\n[3]"},
		{name: "page up moves a page", start: 5, direction: move.PageUp, wantShown: "[2]\n3\n4"},
		{name: "none stays", start: 4, direction: move.None, wantShown: "2\n3\n[4]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := move.Cursor{}.At(tt.start, rows, height).Moved(tt.direction, rows, height)

			assert.Equal(t, tt.wantShown, got.VisibleRows(rows, height, strconv.Itoa, bracketed))
		})
	}
}

func TestCursor_VisibleRows(t *testing.T) {
	t.Parallel()

	t.Run("shows a page ending at the scrolled cursor and marks it", func(t *testing.T) {
		t.Parallel()

		got := move.Cursor{}.At(4, rows, height).VisibleRows(rows, height, strconv.Itoa, bracketed)

		assert.Equal(t, "2\n3\n[4]", got)
	})

	t.Run("stops at the last row", func(t *testing.T) {
		t.Parallel()

		got := move.Cursor{}.At(1, 2, height).VisibleRows(2, height, strconv.Itoa, bracketed)

		assert.Equal(t, "0\n[1]", got)
	})
}

func bracketed(line string) string {
	return "[" + line + "]"
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

func TestCursor_RowAt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		start     int
		rows      int
		line      int
		wantIndex int
		wantOK    bool
	}{
		{name: "finds the row on a line", start: 0, rows: rows, line: 1, wantIndex: 1, wantOK: true},
		{name: "counts from the first visible row", start: 5, rows: rows, line: 0, wantIndex: 3, wantOK: true},
		{name: "finds no row above the first line", start: 0, rows: rows, line: -1, wantIndex: 0, wantOK: false},
		{name: "finds no row below the last line", start: 0, rows: rows, line: height, wantIndex: 0, wantOK: false},
		{name: "finds no row past the last row", start: 0, rows: 2, line: 2, wantIndex: 0, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			index, ok := move.Cursor{}.At(tt.start, tt.rows, height).RowAt(tt.line, tt.rows, height)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantIndex, index)
		})
	}
}

func TestCursor_ClickedAt(t *testing.T) {
	t.Parallel()

	t.Run("moves to the row on the line", func(t *testing.T) {
		t.Parallel()

		got, ok := move.Cursor{}.At(5, rows, height).ClickedAt(0, rows, height)

		assert.True(t, ok)
		assert.Equal(t, 3, got.Index())
	})

	t.Run("stays without a row on the line", func(t *testing.T) {
		t.Parallel()

		got, ok := move.Cursor{}.At(1, 2, height).ClickedAt(2, 2, height)

		assert.False(t, ok)
		assert.Equal(t, 1, got.Index())
	})
}

func TestCursor_Scrolled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		start     int
		lines     int
		wantShown string
	}{
		{name: "scrolls down without moving the cursor", start: 0, lines: 2, wantShown: "2\n3\n4"},
		{name: "scrolls up without moving the cursor", start: rows - 1, lines: -2, wantShown: "5\n6\n7"},
		{name: "stops at the last page", start: 0, lines: rows, wantShown: "7\n8\n9"},
		{name: "stops at the first row", start: 1, lines: -rows, wantShown: "0\n[1]\n2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := move.Cursor{}.At(tt.start, rows, height).Scrolled(tt.lines, rows, height)

			assert.Equal(t, tt.wantShown, got.VisibleRows(rows, height, strconv.Itoa, bracketed))
			assert.Equal(t, tt.start, got.Index())
		})
	}
}
