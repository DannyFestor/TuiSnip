package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type snippetRecord struct {
	ID          string           `json:"id"`
	FolderID    *string          `json:"folder_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	TagIDs      []string         `json:"tag_ids"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Fragments   []fragmentRecord `json:"fragments"`
}

func snippetRecordOf(snippet domain.Snippet) snippetRecord {
	return snippetRecord{
		ID:          snippet.ID().String(),
		FolderID:    folderReference(snippet, domain.Snippet.AtRoot, domain.Snippet.FolderID),
		Title:       snippet.Title().String(),
		Description: snippet.Description().String(),
		TagIDs:      recordsOf(snippet.Tags(), tagIDOf),
		CreatedAt:   timestampOf(snippet.CreatedAt()),
		UpdatedAt:   timestampOf(snippet.UpdatedAt()),
		Fragments:   recordsOf(snippet.Fragments(), fragmentRecordOf),
	}
}
