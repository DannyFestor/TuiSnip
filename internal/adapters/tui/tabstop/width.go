package tabstop

import "github.com/rivo/uniseg"

const (
	tab          = '\t'
	cellsPerStop = 4
)

func IsTab(r rune) bool {
	return r == tab
}

func WidthFrom(runes []rune, column int) int {
	return uniseg.StringWidth(ExpandFrom(runes, column))
}

func RuneCellsAt(r rune, column int) int {
	if IsTab(r) {
		return cellsOfTabAt(column)
	}

	return uniseg.StringWidth(string(r))
}

func cellsOfTabAt(column int) int {
	return cellsPerStop - column%cellsPerStop
}
