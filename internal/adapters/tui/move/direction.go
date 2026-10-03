package move

type Direction int

const (
	None Direction = iota
	Down
	Up
	Top
	Bottom
	PageDown
	PageUp
)
