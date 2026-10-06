package sqlite

import (
	"errors"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func folderFromRow(row sqlcgen.Folder) (domain.Folder, error) {
	name, nameErr := value.NewFolderName(row.Name)
	language, languageErr := value.NewLanguage(row.DefaultLanguage)

	err := errors.Join(nameErr, languageErr)
	if err != nil {
		return domain.Folder{}, corrupt(err)
	}

	folder, err := domain.NewFolder(
		entityID[domain.FolderID](row.ID),
		name,
		folderFromColumn(row.ParentID),
		language,
		time.Time(row.CreatedAt),
		time.Time(row.UpdatedAt),
	)
	if err != nil {
		return domain.Folder{}, corrupt(err)
	}

	return folder, nil
}

func folderIDsOf(rows []sqltype.ID) []domain.FolderID {
	ids := make([]domain.FolderID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, entityID[domain.FolderID](row))
	}

	return ids
}
