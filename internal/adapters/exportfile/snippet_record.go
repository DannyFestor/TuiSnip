package exportfile

import "github.com/DannyFestor/TuiSnip/internal/domain"

type snippetRecord struct {
	snippetFields

	DeletedAt        *string  `json:"deleted_at"`
	FormerFolderPath []string `json:"former_folder_path"`
}

func snippetRecordOf(trash domain.Trash) func(snippet domain.Snippet) snippetRecord {
	return func(snippet domain.Snippet) snippetRecord {
		return snippetRecord{
			snippetFields:    snippetFieldsOf(snippet),
			DeletedAt:        deletedAtOf(trash.Snippets, snippet.ID()),
			FormerFolderPath: formerPathOf(trash.Snippets, snippet.ID()),
		}
	}
}

func snippetBuilder(tags []domain.Tag) func(record snippetRecord) (domain.LibraryEntry[domain.Snippet], error) {
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, tag := range tags {
		tagsByID[tag.ID()] = tag
	}

	return func(record snippetRecord) (domain.LibraryEntry[domain.Snippet], error) {
		return record.entry(tagsByID)
	}
}

func (r snippetRecord) entry(tagsByID map[domain.TagID]domain.Tag) (domain.LibraryEntry[domain.Snippet], error) {
	snippet, err := r.snippet(tagsByID)
	if err != nil {
		return domain.LibraryEntry[domain.Snippet]{}, err
	}

	trash, err := parseTrashFields(r.DeletedAt, r.FormerFolderPath)
	if err != nil {
		return domain.LibraryEntry[domain.Snippet]{}, r.named(err)
	}

	entry, err := domain.NewLibraryEntry(snippet, trash.deletedAt, trash.formerPath)
	if err != nil {
		return domain.LibraryEntry[domain.Snippet]{}, r.named(err)
	}

	return entry, nil
}
