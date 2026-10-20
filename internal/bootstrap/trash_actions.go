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

func newTrashActions(repository *sqlite.TrashRepository) (trashActions, error) {
	list, listErr := trash.NewList(repository)
	restore, restoreErr := trash.NewRestore(repository)

	return trashActions{list: list, restore: restore}, errors.Join(listErr, restoreErr)
}
