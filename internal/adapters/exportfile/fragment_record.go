package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type fragmentRecord struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Language  string `json:"language"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func fragmentRecordOf(fragment domain.Fragment) fragmentRecord {
	return fragmentRecord{
		ID:        fragment.ID().String(),
		Title:     fragment.Title().String(),
		Language:  fragment.Language().String(),
		Content:   fragment.Content().String(),
		CreatedAt: timestampOf(fragment.CreatedAt()),
		UpdatedAt: timestampOf(fragment.UpdatedAt()),
	}
}
