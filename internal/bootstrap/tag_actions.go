package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

type tagActions struct {
	create        *tag.Create
	rename        *tag.Rename
	previewDelete *tag.PreviewDelete
	delete        *tag.Delete
}

func newTagActions(tags *sqlite.TagRepository) (tagActions, error) {
	create, createErr := tag.NewCreate(tags, system.NewIDs(), system.NewClock())
	rename, renameErr := tag.NewRename(tags, system.NewClock())
	previewDelete, previewDeleteErr := tag.NewPreviewDelete(tags)
	deleteAction, deleteErr := tag.NewDelete(tags)

	return tagActions{create: create, rename: rename, previewDelete: previewDelete, delete: deleteAction},
		errors.Join(createErr, renameErr, previewDeleteErr, deleteErr)
}
