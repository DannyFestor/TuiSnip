package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/tag"

type TagDeleteRequested struct {
	Input tag.DeleteInput
}

func (TagDeleteRequested) isOutcome() {}
