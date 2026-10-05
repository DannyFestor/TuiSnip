package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type TagSelected struct {
	ID domain.TagID
}

func (TagSelected) isOutcome() {}
