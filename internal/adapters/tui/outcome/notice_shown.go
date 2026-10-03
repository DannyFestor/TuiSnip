package outcome

type NoticeShown struct {
	Text string
}

func (NoticeShown) isOutcome() {}
