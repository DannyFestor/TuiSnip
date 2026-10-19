package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type snippetActions struct {
	create    *snippet.Create
	update    *snippet.Update
	duplicate *snippet.Duplicate
	remove    *snippet.Delete
	move      *snippet.Move
	capture   *snippet.Capture
	copies    copies
}

func newSnippetActions(
	repos repositories, settings config.Copy, clipboardOptions clipboard.Options,
) (snippetActions, error) {
	create, createErr := snippet.NewCreate(repos.snippets, system.NewIDs(), system.NewClock())
	update, updateErr := snippet.NewUpdate(repos.snippets, system.NewIDs(), system.NewClock())
	duplicate, duplicateErr := snippet.NewDuplicate(repos.snippets, system.NewIDs(), system.NewClock())
	remove, removeErr := snippet.NewDelete(repos.snippets, system.NewClock())
	move, moveErr := snippet.NewMove(repos.snippets)
	capture, captureErr := snippet.NewCapture(clipboard.NewNative(clipboardOptions))
	copied, copiesErr := newCopies(settings, clipboardOptions, repos)

	return snippetActions{
		create:    create,
		update:    update,
		duplicate: duplicate,
		remove:    remove,
		move:      move,
		capture:   capture,
		copies:    copied,
	}, errors.Join(createErr, updateErr, duplicateErr, removeErr, moveErr, captureErr, copiesErr)
}
