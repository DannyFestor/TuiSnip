package tagrefusal

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	blankNameText = "Tag name is blank"
	commaNameText = "Tag name contains a comma"
)

func Text(err error) string {
	switch {
	case errors.Is(err, value.ErrBlankTagName):
		return blankNameText
	case errors.Is(err, value.ErrTagNameHasComma):
		return commaNameText
	}

	return look.FailureText
}
