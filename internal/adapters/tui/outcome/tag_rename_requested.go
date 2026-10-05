package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/tag"

type TagRenameRequested struct {
	Input tag.RenameInput
}

func (TagRenameRequested) isOutcome() {}
