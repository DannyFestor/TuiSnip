package textarea

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"

type drawnSplit struct {
	before string
	inside string
	after  string
}

func splitDrawn(runes []rune, start, end int) drawnSplit {
	insideCell := tabstop.WidthFrom(runes[:start], 0)
	afterCell := insideCell + tabstop.WidthFrom(runes[start:end], insideCell)

	return drawnSplit{
		before: tabstop.ExpandFrom(runes[:start], 0),
		inside: tabstop.ExpandFrom(runes[start:end], insideCell),
		after:  tabstop.ExpandFrom(runes[end:], afterCell),
	}
}
