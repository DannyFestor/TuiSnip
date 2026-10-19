package tabstop

import "github.com/rivo/uniseg"

const (
	tab   = '\t'
	every = 4
)

func WidthFrom(runes []rune, column int) int {
	return uniseg.StringWidth(ExpandFrom(runes, column))
}

func cellsOfTabAt(column int) int {
	return every - column%every
}
