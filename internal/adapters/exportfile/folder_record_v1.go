package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type folderRecordV1 struct {
	ID              string  `json:"id"`
	ParentID        *string `json:"parent_id"`
	Name            string  `json:"name"`
	DefaultLanguage string  `json:"default_language"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func folderRecordV1Of(folder domain.Folder) folderRecordV1 {
	return folderRecordV1{
		ID:              folder.ID().String(),
		ParentID:        folderReference(folder, domain.Folder.AtRoot, domain.Folder.ParentID),
		Name:            folder.Name().String(),
		DefaultLanguage: folder.DefaultLanguage().String(),
		CreatedAt:       timestampOf(folder.CreatedAt()),
		UpdatedAt:       timestampOf(folder.UpdatedAt()),
	}
}

func (r folderRecordV1) upgraded() folderRecord {
	return folderRecord{folderRecordV1: r, DeletedAt: nil, FormerParentPath: []string{}}
}

func (r folderRecordV1) folder() (domain.Folder, error) {
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

func (r folderRecordV1) named(err error) error {
	return fmt.Errorf("folder %s: %w", r.ID, err)
}
