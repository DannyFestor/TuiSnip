package textarea

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
)

type wrappedSegment struct {
	row        int
	index      int
	base       int
	lineLength int
	wrapped    wrappedLine
	colours    []lipgloss.Range
}

func (s wrappedSegment) drawn(from, upTo int, style lipgloss.Style) string {
	var drawn strings.Builder

	runes := s.wrapped.runes
	cell := s.wrapped.startCell + s.wrapped.widthBefore(from)
	written := from

	for index := range s.colours {
		coloured := &s.colours[index]
		start := clamp(coloured.Start-s.base, written, upTo)
		end := clamp(coloured.End-s.base, start, upTo)
		cell = drawRunes(&drawn, runes[written:start], cell, style)
		cell = drawRunes(&drawn, runes[start:end], cell, coloured.Style.Inherit(style))
		written = end
	}

	drawRunes(&drawn, runes[written:upTo], cell, style)

	return drawn.String()
}

func drawRunes(out *strings.Builder, runes []rune, cell int, style lipgloss.Style) int {
	if len(runes) == 0 {
		return cell
	}

	out.WriteString(style.Render(tabstop.ExpandFrom(runes, cell)))

	return cell + tabstop.WidthFrom(runes, cell)
}
