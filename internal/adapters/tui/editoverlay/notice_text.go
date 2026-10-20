package editoverlay

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const pasteTooLarge = "Paste would make Content larger than 256 KiB"

func fieldErrorText(fieldErr domain.FieldError) string {
	switch {
	case errors.Is(fieldErr, value.ErrBlankTitle):
		return "Title is blank"
	case errors.Is(fieldErr, value.ErrTitleTooLong):
		return "Title is longer than 200 characters"
	case errors.Is(fieldErr, value.ErrDescriptionTooLong):
		return "Description is longer than 2000 characters"
	case errors.Is(fieldErr, value.ErrFragmentTitleTooLong):
		return "Fragment title is longer than 200 characters"
	case errors.Is(fieldErr, value.ErrContentTooLong):
		return "Content is larger than 256 KiB"
	}

	return look.FailureText
}
