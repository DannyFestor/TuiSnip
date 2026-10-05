package tagpane

import (
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type takenNameError struct {
	existing value.TagName
}

func (e takenNameError) Error() string {
	return "tagpane: " + e.existing.String() + ": " + domain.ErrTagNameTaken.Error()
}

func (takenNameError) Unwrap() error {
	return domain.ErrTagNameTaken
}
