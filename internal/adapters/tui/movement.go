package tui

type movement int

const (
	moveNone movement = iota
	moveDown
	moveUp
	moveTop
	moveBottom
	movePageDown
	movePageUp
)
