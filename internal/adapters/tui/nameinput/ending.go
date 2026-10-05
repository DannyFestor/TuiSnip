package nameinput

type Ending int

const (
	Typing Ending = iota
	Committed
	Cancelled
	Refused
)
