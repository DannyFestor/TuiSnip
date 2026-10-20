package look

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

type LineColours struct {
	looks []lipgloss.Style
	lines [][]colouredSpan
}

func ColourLines(content, language, codeStyle string) LineColours {
	iterator, err := tokenise(content, language)
	if err != nil {
		return LineColours{looks: nil, lines: nil}
	}

	colouring := newLineColouring(styles.Get(codeStyle))
	for token := iterator(); token != chroma.EOF; token = iterator() {
		colouring.add(token)
	}

	return LineColours{looks: colouring.looks, lines: colouring.lines}
}

func (c LineColours) Highlight(row int, _ []rune) []lipgloss.Range {
	if row >= len(c.lines) {
		return nil
	}

	ranges := make([]lipgloss.Range, 0, len(c.lines[row]))
	for _, span := range c.lines[row] {
		ranges = append(ranges, lipgloss.NewRange(span.start, span.end, c.looks[span.look]))
	}

	return ranges
}

// A lipgloss.Style is over 600 bytes, and 256 KiB of Content holds tens of
// thousands of tokens, so a span points into a table of looks instead.
type colouredSpan struct {
	start int
	end   int
	look  int
}

type lineColouring struct {
	style     *chroma.Style
	looks     []lipgloss.Style
	lookIndex map[chroma.TokenType]int
	lines     [][]colouredSpan
	column    int
}

func newLineColouring(style *chroma.Style) *lineColouring {
	return &lineColouring{
		style:     withoutBackground(style),
		looks:     nil,
		lookIndex: map[chroma.TokenType]int{},
		lines:     [][]colouredSpan{nil},
		column:    0,
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

	entry := c.style.Get(tokenType)
	if runes == 0 || entry.IsZero() {
		return
	}

	row := len(c.lines) - 1
	c.lines[row] = append(c.lines[row], colouredSpan{start: start, end: c.column, look: c.lookOf(tokenType, entry)})
}

func (c *lineColouring) lookOf(tokenType chroma.TokenType, entry chroma.StyleEntry) int {
	if index, ok := c.lookIndex[tokenType]; ok {
		return index
	}

	index := len(c.looks)
	c.looks = append(c.looks, entryLook(entry))
	c.lookIndex[tokenType] = index

	return index
}

func entryLook(entry chroma.StyleEntry) lipgloss.Style {
	look := lipgloss.NewStyle()

	if entry.Bold == chroma.Yes {
		look = look.Bold(true)
	}

	if entry.Underline == chroma.Yes {
		look = look.Underline(true)
	}

	if entry.Italic == chroma.Yes {
		look = look.Italic(true)
	}

	if entry.Colour.IsSet() {
		look = look.Foreground(lipgloss.Color(entry.Colour.String()))
	}

	if entry.Background.IsSet() {
		look = look.Background(lipgloss.Color(entry.Background.String()))
	}

	return look
}

// The Snippet pane's TTY16m formatter clears the style's background before
// colouring, so the textarea drops it too to show the same colours.
func withoutBackground(style *chroma.Style) *chroma.Style {
	builder := style.Builder()
	background := builder.Get(chroma.Background)
	background.Background = 0
	background.NoInherit = true
	builder.AddEntry(chroma.Background, background)

	cleared, err := builder.Build()
	if err != nil {
		return style
	}

	return cleared
}
