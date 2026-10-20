package textarea_test

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/test/keypress"
)

var redText = regexp.MustCompile(`\x1b\[(?:[0-9;]*;)?38;2;255;0;0(?:;[0-9;]*)?m([^\x1b]*)`)

type fixedColours map[int][]lipgloss.Range

func (f fixedColours) Highlight(row int, _ []rune) []lipgloss.Range {
	return f[row]
}

func TestModel_View_highlighted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		width   int
		colours fixedColours
		want    []string
	}{
		{
			name:    "colours the runes a range covers",
			content: "let x",
			width:   20,
			colours: fixedColours{0: {red(0, 3)}},
			want:    []string{"let"},
		},
		{
			name:    "colours the logical line the range belongs to",
			content: "one\ntwo",
			width:   20,
			colours: fixedColours{1: {red(1, 3)}},
			want:    []string{"wo"},
		},
		{
			name:    "counts a range on a wrapped row from the start of the line",
			content: wrappedTabLine,
			width:   10,
			colours: fixedColours{0: {red(10, 12)}},
			want:    []string{"jk"},
		},
		{
			name:    "colours each row a range runs across",
			content: wrappedTabLine,
			width:   10,
			colours: fixedColours{0: {red(6, 12)}},
			want:    []string{"ghi", "jk"},
		},
		{
			name:    "draws a tab inside a range up to the next tab stop",
			content: "a\tb",
			width:   20,
			colours: fixedColours{0: {red(0, 3)}},
			want:    []string{"a   b"},
		},
		{
			name:    "clips a range that runs past the end of the line",
			content: "ab\ncd",
			width:   20,
			colours: fixedColours{0: {red(1, 40)}},
			want:    []string{"b"},
		},
		{
			name:    "skips a range that starts past the end of the line",
			content: "ab\ncd",
			width:   20,
			colours: fixedColours{0: {red(5, 9)}},
			want:    nil,
		},
		{
			name:    "colours only the part of a range after an earlier one it overlaps",
			content: "abcd",
			width:   20,
			colours: fixedColours{0: {red(0, 2), red(1, 3)}},
			want:    []string{"ab", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			area := newPlainTextarea(t, tt.content)
			area.SetWidth(tt.width)
			area.SetHighlighter(tt.colours)

			assert.Equal(t, tt.want, colouredRed(area.View()))
		})
	}

	t.Run("leaves the cell under the cursor to the cursor", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {red(0, 3)}})
		area.SetCursorColumn(1)

		assert.Equal(t, []string{"a", "c"}, colouredRed(area.View()))
	})

	t.Run("leaves selected text to the selection", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {red(0, 3)}})
		area.SetCursorColumn(1)
		area, _ = area.Update(keypress.Shift(tea.KeyRight))

		assert.Equal(t, []string{"a", "c"}, colouredRed(area.View()))
	})

	t.Run("draws the text uncoloured without a highlighter", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abc")
		area.SetHighlighter(fixedColours{0: {red(0, 3)}})
		area.SetHighlighter(nil)

		assert.Empty(t, colouredRed(area.View()))
	})
}

func red(start, end int) lipgloss.Range {
	return lipgloss.NewRange(start, end, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")))
}

func colouredRed(view string) []string {
	var coloured []string

	for _, match := range redText.FindAllStringSubmatch(view, -1) {
		if text := strings.TrimRight(match[1], " "); text != "" {
			coloured = append(coloured, text)
		}
	}

	return coloured
}
