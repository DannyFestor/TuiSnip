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
	previewDelete      *folder.PreviewDelete
	remove             *folder.Delete
	setDefaultLanguage *folder.SetDefaultLanguage
}

func newFolderActions(folders *sqlite.FolderRepository) (folderActions, error) {
	create, createErr := folder.NewCreate(folders, system.NewIDs(), system.NewClock())
	rename, renameErr := folder.NewRename(folders, system.NewClock())
	previewDelete, previewDeleteErr := folder.NewPreviewDelete(folders)
	remove, removeErr := folder.NewDelete(folders)
	setDefaultLanguage, setDefaultLanguageErr := folder.NewSetDefaultLanguage(folders, system.NewClock())

	return folderActions{
		create:             create,
		rename:             rename,
		previewDelete:      previewDelete,
		remove:             remove,
		setDefaultLanguage: setDefaultLanguage,
	}, errors.Join(createErr, renameErr, previewDeleteErr, removeErr, setDefaultLanguageErr)
}
