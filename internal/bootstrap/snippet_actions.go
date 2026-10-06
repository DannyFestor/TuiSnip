package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type snippetActions struct {
	create    *snippet.Create
	update    *snippet.Update
	duplicate *snippet.Duplicate
	remove    *snippet.Delete
}

func newSnippetActions(snippets *sqlite.SnippetRepository) (snippetActions, error) {
	create, createErr := snippet.NewCreate(snippets, system.NewIDs(), system.NewClock())
	update, updateErr := snippet.NewUpdate(snippets, system.NewIDs(), system.NewClock())
	duplicate, duplicateErr := snippet.NewDuplicate(snippets, system.NewIDs(), system.NewClock())
	remove, removeErr := snippet.NewDelete(snippets)

	return snippetActions{
		create:    create,
		update:    update,
		duplicate: duplicate,
		remove:    remove,
	}, errors.Join(createErr, updateErr, duplicateErr, removeErr)
}
