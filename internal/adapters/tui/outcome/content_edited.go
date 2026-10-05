package outcome

type ContentEdited struct {
	Asked   ExternalEditAsked
	Content string
}

func (ContentEdited) isOutcome() {}
