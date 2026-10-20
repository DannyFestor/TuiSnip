package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
)

type folderActions struct {
	create             *folder.Create
	rename             *folder.Rename
	remove             *folder.Delete
	setDefaultLanguage *folder.SetDefaultLanguage
	move               *folder.Move
}

func newFolderActions(folders *sqlite.FolderRepository) (folderActions, error) {
	create, createErr := folder.NewCreate(folders, system.NewIDs(), system.NewClock())
	rename, renameErr := folder.NewRename(folders, system.NewClock())
	remove, removeErr := folder.NewDelete(folders, system.NewClock())
	setDefaultLanguage, setDefaultLanguageErr := folder.NewSetDefaultLanguage(folders, system.NewClock())
	move, moveErr := folder.NewMove(folders)

	return folderActions{
		create:             create,
		rename:             rename,
		remove:             remove,
		setDefaultLanguage: setDefaultLanguage,
		move:               move,
	}, errors.Join(createErr, renameErr, removeErr, setDefaultLanguageErr, moveErr)
}
