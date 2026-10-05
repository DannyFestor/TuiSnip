package tui

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	copiedText            = "Copied"
	sentToTerminalText    = "Sent to terminal"
	noClipboardText       = "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)"
	clipboardEmptyText    = "Clipboard is empty"
	clipboardTooLargeText = "Clipboard is larger than 256 KiB"
	noCaptureToolText     = "No clipboard tool found (pbpaste, wl-paste, xclip, xsel)"
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

func captureRefusalText(err error) (string, bool) {
	switch {
	case errors.Is(err, domain.ErrClipboardEmpty):
		return clipboardEmptyText, true
	case errors.Is(err, value.ErrContentTooLong):
		return clipboardTooLargeText, true
	}

	return "", false
}

func captureFailureText(err error) string {
	if errors.Is(err, domain.ErrNoClipboardTool) {
		return noCaptureToolText
	}

	return look.FailureText
}
