package outcome

type TrashCopyRequested CopyRequested

func (TrashCopyRequested) isOutcome() {}
