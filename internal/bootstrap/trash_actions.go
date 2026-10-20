package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

type trashActions struct {
	list              *trash.List
	restore           *trash.Restore
	deletePermanently *trash.DeletePermanently
	empty             *trash.Empty
}

func newTrashActions(repository *sqlite.TrashRepository) (trashActions, error) {
	list, listErr := trash.NewList(repository)
	restore, restoreErr := trash.NewRestore(repository)
	deletePermanently, deleteErr := trash.NewDeletePermanently(repository)
	empty, emptyErr := trash.NewEmpty(repository)

	return trashActions{list: list, restore: restore, deletePermanently: deletePermanently, empty: empty},
		errors.Join(listErr, restoreErr, deleteErr, emptyErr)
}
