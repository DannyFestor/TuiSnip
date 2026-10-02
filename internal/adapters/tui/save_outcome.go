package tui

type saveOutcome int

const (
	saveSucceeded saveOutcome = iota
	saveRejected
	saveFailed
)
