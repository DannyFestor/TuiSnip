package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type snippetRecord struct {
	snippetRecordV1

	DeletedAt        *string  `json:"deleted_at"`
	FormerFolderPath []string `json:"former_folder_path"`
}

func snippetRecordOf(trash domain.Trash) func(snippet domain.Snippet) snippetRecord {
	return func(snippet domain.Snippet) snippetRecord {
		return snippetRecord{
			snippetRecordV1:  snippetRecordV1Of(snippet),
			DeletedAt:        deletedAtOf(trash.Snippets, snippet.ID()),
			FormerFolderPath: formerPathOf(trash.Snippets, snippet.ID()),
		}
	}
}

func snippetBuilder(tags []domain.Tag) func(record snippetRecord) (trashEntry[domain.Snippet], error) {
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, tag := range tags {
		tagsByID[tag.ID()] = tag
	}

	return func(record snippetRecord) (trashEntry[domain.Snippet], error) {
		return record.entry(tagsByID)
	}
}

func (r snippetRecord) entry(tagsByID map[domain.TagID]domain.Tag) (trashEntry[domain.Snippet], error) {
	snippet, err := r.snippet(tagsByID)
	if err != nil {
		return trashEntry[domain.Snippet]{}, err
	}

	entry, err := entryOf(snippet, r.DeletedAt, r.FormerFolderPath)
	if err != nil {
		return trashEntry[domain.Snippet]{}, r.named(err)
	}

	return entry, nil
}
