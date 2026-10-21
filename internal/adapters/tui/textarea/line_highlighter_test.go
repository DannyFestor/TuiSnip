package textarea_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencolour"
)

const red = "255;0;0"

type fixedColours map[int][]lipgloss.Range

func (f fixedColours) Highlight(row int, _ []rune) []lipgloss.Range {
	return f[row]
}

func TestModel_View_highlighted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		width    int
		colours  fixedColours
		coloured []string
		plain    []string
	}{
		{
			name:     "colours the runes a range covers",
			content:  "let x",
			width:    20,
			colours:  fixedColours{0: {redRange(0, 3)}},
			coloured: []string{"let"},
			plain:    []string{"x"},
		},
		{
			name:     "colours the logical line the range belongs to",
			content:  "one\ntwo",
			width:    20,
			colours:  fixedColours{1: {redRange(1, 3)}},
			coloured: []string{"wo"},
			plain:    []string{"one"},
		},
		{
			name:     "counts a range on a wrapped row from the start of the line",
			content:  wrappedTabLine,
			width:    10,
			colours:  fixedColours{0: {redRange(10, 12)}},
			coloured: []string{"jk"},
			plain:    []string{"abcdefghi"},
		},
		{
			name:     "colours each row a range runs across",
			content:  wrappedTabLine,
			width:    10,
			colours:  fixedColours{0: {redRange(6, 12)}},
			coloured: []string{"ghi", "jk"},
			plain:    []string{"abcdef"},
		},
		{
			name:     "draws a tab inside a range up to the next tab stop",
			content:  "a\tb",
			width:    20,
			colours:  fixedColours{0: {redRange(0, 3)}},
			coloured: []string{"a   b"},
			plain:    nil,
		},
		{
			name:     "clips a range that runs past the end of the line",
			content:  "ab\ncd",
			width:    20,
			colours:  fixedColours{0: {redRange(1, 40)}},
			coloured: []string{"b"},
			plain:    []string{"a", "cd"},
		},
		{
			name:     "skips a range that starts past the end of the line",
			content:  "ab\ncd",
			width:    20,
			colours:  fixedColours{0: {redRange(5, 9)}},
			coloured: nil,
			plain:    []string{"ab", "cd"},
		},
		{
			name:     "colours only the part of a range after an earlier one it overlaps",
			content:  "abcd",
			width:    20,
			colours:  fixedColours{0: {redRange(0, 2), redRange(1, 3)}},
			coloured: []string{"ab", "c"},
			plain:    []string{"d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			area := newPlainTextarea(t, tt.content)
			area.SetWidth(tt.width)
			area.SetHighlighter(tt.colours)

			assertColoured(t, area, tt.coloured, tt.plain)
		})
	}

	t.Run("leaves the cell under the cursor to the cursor", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {redRange(0, 3)}})
		area.SetCursorColumn(1)

		assertColoured(t, area, []string{"a", "c"}, []string{"b"})
	})

	t.Run("leaves selected text to the selection", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {redRange(0, 3)}})
		area.SetCursorColumn(1)
		area, _ = area.Update(keypress.Shift(tea.KeyRight))

		assertColoured(t, area, []string{"a", "c"}, []string{"b"})
	})

	t.Run("draws the text uncoloured without a highlighter", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {redRange(0, 3)}})
		area.SetHighlighter(nil)

		assertColoured(t, area, nil, []string{"abc"})
	})
}

func redRange(start, end int) lipgloss.Range {
	return lipgloss.NewRange(start, end, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")))
}

func assertColoured(t *testing.T, area textarea.Model, coloured, plain []string) {
	t.Helper()

	view := area.View()

	for _, text := range coloured {
		assert.Equal(t, red, screencolour.Foreground(t, view, text), text)
	}

	for _, text := range plain {
		assert.Empty(t, screencolour.Foreground(t, view, text), text)
	}
}
