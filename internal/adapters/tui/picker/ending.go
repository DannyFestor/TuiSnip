package picker

type Ending int

const (
	Filtering Ending = iota
	Picked
	Cancelled
)
