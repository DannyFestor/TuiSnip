package outcome

type SaveFailed struct {
	Err error
}

func (SaveFailed) isOutcome() {}
