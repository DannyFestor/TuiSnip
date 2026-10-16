package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type folderFields struct {
	ID              string  `json:"id"`
	ParentID        *string `json:"parent_id"`
	Name            string  `json:"name"`
	DefaultLanguage string  `json:"default_language"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func folderFieldsOf(folder domain.Folder) folderFields {
	return folderFields{
		ID:              folder.ID().String(),
		ParentID:        folderReference(folder, domain.Folder.AtRoot, domain.Folder.ParentID),
		Name:            folder.Name().String(),
		DefaultLanguage: folder.DefaultLanguage().String(),
		CreatedAt:       timestampOf(folder.CreatedAt()),
		UpdatedAt:       timestampOf(folder.UpdatedAt()),
	}
}

func (f folderFields) folder() (domain.Folder, error) {
	id, idErr := parseID[domain.FolderID](f.ID)
	parentID, parentErr := parseFolderReference(f.ParentID)
	name, nameErr := value.NewFolderName(f.Name)
	language, languageErr := value.NewLanguage(f.DefaultLanguage)
	stamped, datesErr := parseDates(f.CreatedAt, f.UpdatedAt)

	err := errors.Join(idErr, parentErr, nameErr, languageErr, datesErr)
	if err != nil {
		return domain.Folder{}, f.named(err)
	}

	folder, err := domain.NewFolder(id, name, parentID, language, stamped.created, stamped.updated)
	if err != nil {
		return domain.Folder{}, f.named(err)
	}

	return folder, nil
}

func (f folderFields) named(err error) error {
	return fmt.Errorf("folder %s: %w", f.ID, err)
}
