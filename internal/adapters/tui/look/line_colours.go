package look

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

type LineColours struct {
	tokenStyles []lipgloss.Style
	lines       [][]colouredSpan
}

func ColourLines(content, language, codeStyle string) LineColours {
	iterator, err := tokenise(content, language)
	if err != nil {
		return LineColours{tokenStyles: nil, lines: nil}
	}

	colouring := newLineColouring(styles.Get(codeStyle))
	for token := iterator(); token != chroma.EOF; token = iterator() {
		colouring.add(token)
	}

	return LineColours{tokenStyles: colouring.tokenStyles, lines: colouring.lines}
}

func (c LineColours) Highlight(row int, _ []rune) []lipgloss.Range {
	if row >= len(c.lines) {
		return nil
	}

	ranges := make([]lipgloss.Range, 0, len(c.lines[row]))
	for _, span := range c.lines[row] {
		ranges = append(ranges, lipgloss.NewRange(span.start, span.end, c.tokenStyles[span.tokenStyle]))
	}

	return ranges
}

// A lipgloss.Style is over 600 bytes, and 256 KiB of Content holds tens of
// thousands of tokens, so a span points into a table of styles instead.
type colouredSpan struct {
	start      int
	end        int
	tokenStyle int
}

type lineColouring struct {
	codeStyle   *chroma.Style
	tokenStyles []lipgloss.Style
	styleIndex  map[chroma.TokenType]int
	lines       [][]colouredSpan
	column      int
}

func newLineColouring(codeStyle *chroma.Style) *lineColouring {
	return &lineColouring{
		codeStyle:   withoutBackground(codeStyle),
		tokenStyles: nil,
		styleIndex:  map[chroma.TokenType]int{},
		lines:       [][]colouredSpan{nil},
		column:      0,
	}
}

func (c *lineColouring) add(token chroma.Token) {
	for index, part := range strings.Split(token.Value, "\n") {
		if index > 0 {
			c.lines = append(c.lines, nil)
			c.column = 0
		}

		c.colour(token.Type, utf8.RuneCountInString(part))
	}
}

func (c *lineColouring) colour(tokenType chroma.TokenType, runes int) {
	start := c.column
	c.column += runes

	entry := c.codeStyle.Get(tokenType)
	if runes == 0 || entry.IsZero() {
		return
	}

	row := len(c.lines) - 1
	c.lines[row] = append(c.lines[row], colouredSpan{
		start:      start,
		end:        c.column,
		tokenStyle: c.styleIndexOf(tokenType, entry),
	})
}

func (c *lineColouring) styleIndexOf(tokenType chroma.TokenType, entry chroma.StyleEntry) int {
	if index, ok := c.styleIndex[tokenType]; ok {
		return index
	}

	index := len(c.tokenStyles)
	c.tokenStyles = append(c.tokenStyles, entryStyle(entry))
	c.styleIndex[tokenType] = index

	return index
}

func entryStyle(entry chroma.StyleEntry) lipgloss.Style {
	style := lipgloss.NewStyle()

	if entry.Bold == chroma.Yes {
		style = style.Bold(true)
	}

	if entry.Underline == chroma.Yes {
		style = style.Underline(true)
	}

	if entry.Italic == chroma.Yes {
		style = style.Italic(true)
	}

	if entry.Colour.IsSet() {
		style = style.Foreground(lipgloss.Color(entry.Colour.String()))
	}

	if entry.Background.IsSet() {
		style = style.Background(lipgloss.Color(entry.Background.String()))
	}

	return style
}

// The Snippet pane's TTY16m formatter clears the style's background before
// colouring, so the textarea drops it too to show the same colours.
func withoutBackground(codeStyle *chroma.Style) *chroma.Style {
	builder := codeStyle.Builder()
	background := builder.Get(chroma.Background)
	background.Background = 0
	background.NoInherit = true
	builder.AddEntry(chroma.Background, background)

	cleared, err := builder.Build()
	if err != nil {
		return codeStyle
	}

	return cleared
}
