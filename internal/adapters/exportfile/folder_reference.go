package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

func folderReference[E any](entity E, atRoot func(entity E) bool, folderID func(entity E) domain.FolderID) *string {
	if atRoot(entity) {
		return nil
	}

	return new(folderID(entity).String())
}
