package overlay

type Parent interface {
	Overlay
	Received(outcome Outcome) Step
}
