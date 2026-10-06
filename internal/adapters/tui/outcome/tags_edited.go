package outcome

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"

type TagsEdited struct {
	Chosen tagchoice.Chosen
}

func (TagsEdited) isOutcome() {}
