package outcome

type QuitConfirmed struct{}

func (QuitConfirmed) isOutcome() {}
