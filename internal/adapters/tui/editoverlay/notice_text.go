package editoverlay

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	pasteHasTabs          = "Pasted text contains tabs"
	pasteOverflowsContent = "Paste would make Content longer than 10,000 lines"
)

func refusedPasteText(refusal, externalEditorKey string) string {
	if externalEditorKey == "" {
		return refusal
	}

	return refusal + "; use " + externalEditorKey + " to edit in $EDITOR"
}

func fieldErrorText(fieldErr domain.FieldError) string {
	switch {
	case errors.Is(fieldErr, value.ErrBlankTitle):
		return "Title is blank"
	case errors.Is(fieldErr, value.ErrTitleTooLong):
		return "Title is longer than 200 characters"
	case errors.Is(fieldErr, value.ErrDescriptionTooLong):
		return "Description is longer than 2000 characters"
	case errors.Is(fieldErr, value.ErrContentTooLong):
		return "Content is larger than 256 KiB"
	}

	return look.FailureText
}
