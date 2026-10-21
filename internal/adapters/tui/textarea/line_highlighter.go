package textarea

import "charm.land/lipgloss/v2"

type LineHighlighter interface {
	Highlight(row int, line []rune) []lipgloss.Range
}
