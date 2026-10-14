package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type tagRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func tagRecordOf(tag domain.Tag) tagRecord {
	return tagRecord{
		ID:        tag.ID().String(),
		Name:      tag.Name().String(),
		CreatedAt: timestampOf(tag.CreatedAt()),
		UpdatedAt: timestampOf(tag.UpdatedAt()),
	}
}

func tagIDOf(tag domain.Tag) string {
	return tag.ID().String()
}
