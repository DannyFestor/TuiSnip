package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/tag"

type TagCreateRequested struct {
	Input tag.CreateInput
}

func (TagCreateRequested) isOutcome() {}
