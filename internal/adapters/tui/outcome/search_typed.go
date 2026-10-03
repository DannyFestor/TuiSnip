package outcome

type SearchTyped struct {
	Text string
}

func (SearchTyped) isOutcome() {}
