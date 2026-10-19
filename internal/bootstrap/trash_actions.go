package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

type trashActions struct {
	list    *trash.List
	restore *trash.Restore
}

func newTrashActions(items *sqlite.TrashRepository) (trashActions, error) {
	list, listErr := trash.NewList(items)
	restore, restoreErr := trash.NewRestore(items)

	return trashActions{list: list, restore: restore}, errors.Join(listErr, restoreErr)
}
