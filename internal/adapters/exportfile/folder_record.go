package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type folderRecord struct {
	ID              string  `json:"id"`
	ParentID        *string `json:"parent_id"`
	Name            string  `json:"name"`
	DefaultLanguage string  `json:"default_language"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func folderRecordOf(folder domain.Folder) folderRecord {
	return folderRecord{
		ID:              folder.ID().String(),
		ParentID:        folderReference(folder, domain.Folder.AtRoot, domain.Folder.ParentID),
		Name:            folder.Name().String(),
		DefaultLanguage: folder.DefaultLanguage().String(),
		CreatedAt:       timestampOf(folder.CreatedAt()),
		UpdatedAt:       timestampOf(folder.UpdatedAt()),
	}
}

func (r folderRecord) folder() (domain.Folder, error) {
	id, idErr := parseID[domain.FolderID](r.ID)
	parentID, parentErr := parseFolderReference(r.ParentID)
	name, nameErr := value.NewFolderName(r.Name)
	language, languageErr := value.NewLanguage(r.DefaultLanguage)
	stamped, datesErr := parseDates(r.CreatedAt, r.UpdatedAt)

	err := errors.Join(idErr, parentErr, nameErr, languageErr, datesErr)
	if err != nil {
		return domain.Folder{}, r.named(err)
	}

	folder, err := domain.NewFolder(id, name, parentID, language, stamped.created, stamped.updated)
	if err != nil {
		return domain.Folder{}, r.named(err)
	}

	return folder, nil
}

func (r folderRecord) named(err error) error {
	return fmt.Errorf("folder %s: %w", r.ID, err)
}
