package outcome

type QuitAsked struct{}

func (QuitAsked) isOutcome() {}
