package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type snippetRecordV1 struct {
	ID          string           `json:"id"`
	FolderID    *string          `json:"folder_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	TagIDs      []string         `json:"tag_ids"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Fragments   []fragmentRecord `json:"fragments"`
}

func snippetRecordV1Of(snippet domain.Snippet) snippetRecordV1 {
	return snippetRecordV1{
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

func (r snippetRecordV1) upgraded() snippetRecord {
	return snippetRecord{snippetRecordV1: r, DeletedAt: nil, FormerFolderPath: []string{}}
}

func (r snippetRecordV1) snippet(tagsByID map[domain.TagID]domain.Tag) (domain.Snippet, error) {
	id, idErr := parseID[domain.SnippetID](r.ID)
	folderID, folderErr := parseFolderReference(r.FolderID)
	title, titleErr := value.NewTitle(r.Title)
	description, descriptionErr := value.NewDescription(r.Description)
	fragments, fragmentsErr := entitiesOf(r.Fragments, fragmentRecord.fragment)
	tags, tagsErr := entitiesOf(r.TagIDs, carriedTag(tagsByID))
	stamped, datesErr := parseDates(r.CreatedAt, r.UpdatedAt)

	err := errors.Join(idErr, folderErr, titleErr, descriptionErr, fragmentsErr, tagsErr, datesErr)
	if err != nil {
		return domain.Snippet{}, r.named(err)
	}

	snippet, err := domain.NewSnippet(
		id, title, description, folderID, fragments, tags, stamped.created, stamped.updated,
	)
	if err != nil {
		return domain.Snippet{}, r.named(err)
	}

	return snippet, nil
}

func (r snippetRecordV1) named(err error) error {
	return fmt.Errorf("snippet %s: %w", r.ID, err)
}

func carriedTag(tagsByID map[domain.TagID]domain.Tag) func(rawID string) (domain.Tag, error) {
	return func(rawID string) (domain.Tag, error) {
		id, err := parseID[domain.TagID](rawID)
		if err != nil {
			return domain.Tag{}, fmt.Errorf("tag %s: %w", rawID, err)
		}

		tag, known := tagsByID[id]
		if !known {
			return domain.Tag{}, fmt.Errorf("tag %s: %w", rawID, domain.ErrMissingTag)
		}

		return tag, nil
	}
}
