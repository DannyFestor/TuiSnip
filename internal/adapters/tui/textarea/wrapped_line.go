package textarea

import (
	"slices"
	"unicode"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
)

type wrappedLine struct {
	runes []rune
	// Tab stops count from the start of the whole line, not of the row, so
	// a tab covers the same cells as where the Snippet pane draws it.
	startCell int
}

func (w wrappedLine) width() int {
	return tabstop.WidthFrom(w.runes, w.startCell)
}

func (w wrappedLine) widthBefore(index int) int {
	return tabstop.WidthFrom(w.runes[:index], w.startCell)
}

func (w wrappedLine) widthWith(tail ...[]rune) int {
	return tabstop.WidthFrom(slices.Concat(append([][]rune{w.runes}, tail...)...), w.startCell)
}

func (w wrappedLine) following() wrappedLine {
	return wrappedLine{runes: []rune{}, startCell: w.startCell + w.width()}
}

func (w wrappedLine) withoutTrailingWhitespace() wrappedLine {
	if len(w.runes) > 0 && unicode.IsSpace(w.runes[len(w.runes)-1]) {
		w.runes = w.runes[:len(w.runes)-1]
	}

	return w
}

func (w wrappedLine) drawnBetween(start, end int) string {
	return tabstop.ExpandFrom(w.runes[start:end], w.startCell+w.widthBefore(start))
}

func (w wrappedLine) runeIndexAt(cell int) int {
	if cell <= 0 {
		return 0
	}

	end := w.startCell

	for index, r := range w.runes {
		end += tabstop.RuneCellsAt(r, end)
		if end-w.startCell > cell {
			return index
		}
	}

	return len(w.runes)
}
