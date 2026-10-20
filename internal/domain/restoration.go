package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

//go:generate go-enum

// ENUM(none, folder_in_trash, folder_deleted).
type RestoreFallback string

type Restoration struct {
	FolderID FolderID
	Fallback RestoreFallback
}

func RestorationOf(remembered FolderID, formerPath value.FolderPath, rememberedLive bool) Restoration {
	switch {
	case !formerPath.IsEmpty():
		return Restoration{FolderID: FolderID{}, Fallback: RestoreFallbackFolderDeleted}
	case remembered.IsNil(), rememberedLive:
		return Restoration{FolderID: remembered, Fallback: RestoreFallbackNone}
	}

	return Restoration{FolderID: FolderID{}, Fallback: RestoreFallbackFolderInTrash}
}
