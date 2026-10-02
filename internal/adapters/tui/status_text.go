package tui

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	copiedText         = "Copied"
	sentToTerminalText = "Sent to terminal"
	noClipboardText    = "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)"
	genericFailureText = "Something went wrong; see the log"
	pasteHasTabsText   = "Pasted text contains tabs; use ctrl+e to edit in $EDITOR"
)

var errUnknownDelivery = errors.New("tui: unknown copy delivery")

func deliveryText(delivery domain.CopyDelivery) (string, error) {
	switch delivery {
	case domain.CopyDeliveryPlaced:
		return copiedText, nil
	case domain.CopyDeliverySentToTerminal:
		return sentToTerminalText, nil
	}

	return "", fmt.Errorf("%w %q", errUnknownDelivery, delivery)
}

func failureText(err error) string {
	if errors.Is(err, domain.ErrNoClipboardTool) {
		return noClipboardText
	}

	return genericFailureText
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

	return genericFailureText
}
