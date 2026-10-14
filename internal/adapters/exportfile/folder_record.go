package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

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
