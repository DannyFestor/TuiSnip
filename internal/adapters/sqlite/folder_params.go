package sqlite

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func insertFolderParams(folder domain.Folder) sqlcgen.InsertFolderParams {
	return sqlcgen.InsertFolderParams{
		ID:              columnID(folder.ID()),
		ParentID:        folderColumn(folder.ParentID()),
		Name:            folder.Name().String(),
		DefaultLanguage: folder.DefaultLanguage().String(),
		CreatedAt:       sqltype.Timestamp(folder.CreatedAt()),
		UpdatedAt:       sqltype.Timestamp(folder.UpdatedAt()),
	}
}
