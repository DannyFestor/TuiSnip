package tui

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	copiedText         = "Copied"
	sentToTerminalText = "Sent to terminal"
	noClipboardText    = "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)"
	clipboardEmptyText = "Clipboard is empty"
	noCaptureToolText  = "No clipboard tool found (pbpaste, wl-paste, xclip, xsel)"
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

	return look.FailureText
}

func captureFailureText(err error) string {
	if errors.Is(err, domain.ErrNoClipboardTool) {
		return noCaptureToolText
	}

	return look.FailureText
}
