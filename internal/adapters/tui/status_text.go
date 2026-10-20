package tui

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
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
	noEditorText          = "No editor found; set `editor` in config.toml"
	editorFailedText      = "Editor exited with an error; changes discarded"
)

const (
	snippetNoun        = "Snippet"
	folderNoun         = "Folder"
	movedToTrashFormat = "Moved %s %q to the Trash"
)

const (
	restoredFormat       = "Restored %q to %s"
	restoredToRootFormat = "Restored %q to the Root"
	homeInTrashFormat    = "; its %s is in the Trash"
	homeDeletedFormat    = "; its %s was deleted"
	parentPrefix         = "parent "
)

var errUnknownDelivery = errors.New("tui: unknown copy delivery")

func movedToTrashText(noun, name string) string {
	return fmt.Sprintf(movedToTrashFormat, noun, name)
}

func restoredText(requested outcome.TrashRestoreRequested, restoration domain.Restoration) string {
	toRoot := fmt.Sprintf(restoredToRootFormat, requested.Name)

	switch restoration.Fallback {
	case domain.RestoreFallbackFolderInTrash:
		return toRoot + fmt.Sprintf(homeInTrashFormat, homeNoun(requested.Input.Item))
	case domain.RestoreFallbackFolderDeleted:
		return toRoot + fmt.Sprintf(homeDeletedFormat, homeNoun(requested.Input.Item))
	case domain.RestoreFallbackNone:
	}

	if restoration.FolderID.IsNil() {
		return toRoot
	}

	return fmt.Sprintf(restoredFormat, requested.Name, folderpath.ShortOf(requested.Path))
}

func homeNoun(restored trash.ItemID) string {
	if restored.Kind() == trash.ItemKindFolder {
		return parentPrefix + folderNoun
	}

	return folderNoun
}

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

func externalEditRefusalText(err error) (string, bool) {
	switch {
	case errors.Is(err, domain.ErrNoEditor):
		return noEditorText, true
	case errors.Is(err, domain.ErrEditorFailed):
		return editorFailedText, true
	}

	return "", false
}

func externalEditFailureText(err error) string {
	if startErr, ok := errors.AsType[domain.EditorStartError](err); ok {
		return startErr.Err.Error()
	}

	return look.FailureText
}
