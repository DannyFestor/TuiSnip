package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type TagDeleteAsked struct {
	ID domain.TagID
}

func (TagDeleteAsked) isOutcome() {}
