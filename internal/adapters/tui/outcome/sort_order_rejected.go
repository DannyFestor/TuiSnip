package outcome

type SortOrderRejected struct {
	Err error
}

func (SortOrderRejected) isOutcome() {}
