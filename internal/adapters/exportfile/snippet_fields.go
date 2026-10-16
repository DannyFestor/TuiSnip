package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type snippetFields struct {
	ID          string           `json:"id"`
	FolderID    *string          `json:"folder_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	TagIDs      []string         `json:"tag_ids"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Fragments   []fragmentRecord `json:"fragments"`
}

func snippetFieldsOf(snippet domain.Snippet) snippetFields {
	return snippetFields{
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

func (f snippetFields) snippet(tagsByID map[domain.TagID]domain.Tag) (domain.Snippet, error) {
	id, idErr := parseID[domain.SnippetID](f.ID)
	folderID, folderErr := parseFolderReference(f.FolderID)
	title, titleErr := value.NewTitle(f.Title)
	description, descriptionErr := value.NewDescription(f.Description)
	fragments, fragmentsErr := entitiesOf(f.Fragments, fragmentRecord.fragment)
	tags, tagsErr := entitiesOf(f.TagIDs, carriedTag(tagsByID))
	stamped, datesErr := parseDates(f.CreatedAt, f.UpdatedAt)

	err := errors.Join(idErr, folderErr, titleErr, descriptionErr, fragmentsErr, tagsErr, datesErr)
	if err != nil {
		return domain.Snippet{}, f.named(err)
	}

	snippet, err := domain.NewSnippet(
		id, title, description, folderID, fragments, tags, stamped.created, stamped.updated,
	)
	if err != nil {
		return domain.Snippet{}, f.named(err)
	}

	return snippet, nil
}

func (f snippetFields) named(err error) error {
	return fmt.Errorf("snippet %s: %w", f.ID, err)
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
