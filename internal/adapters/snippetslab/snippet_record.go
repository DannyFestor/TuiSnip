package snippetslab

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type snippetRecord struct {
	UUID           string           `json:"uuid"`
	Title          string           `json:"title"`
	Folder         string           `json:"folder"`
	Tags           []string         `json:"tags"`
	Pinned         bool             `json:"pinned"`
	GistIdentifier string           `json:"gistIdentifier"`
	DateCreated    string           `json:"dateCreated"`
	DateModified   string           `json:"dateModified"`
	DateDeleted    string           `json:"dateDeleted"`
	Fragments      []fragmentRecord `json:"fragments"`
}

func (c *conversion) snippetEntry(record snippetRecord) (domain.LibraryEntry[domain.Snippet], error) {
	snippet, snippetErr := c.snippet(record)
	deletedAt, deletedErr := parseDeletedDate(record.DateDeleted)

	err := errors.Join(snippetErr, deletedErr)
	if err != nil {
		return domain.LibraryEntry[domain.Snippet]{}, record.named(err)
	}

	entry, err := domain.NewLibraryEntry(snippet, deletedAt, value.FolderPath{})
	if err != nil {
		return domain.LibraryEntry[domain.Snippet]{}, record.named(err)
	}

	return entry, nil
}

func (c *conversion) snippet(record snippetRecord) (domain.Snippet, error) {
	title, titleErr := value.NewTitle(record.Title)
	fragments, fragmentsErr := c.fragments(title, record.Fragments)
	description, descriptionErr := descriptionOf(record.Fragments)
	stamped, datesErr := parseDates(record.DateCreated, record.DateModified)

	err := errors.Join(titleErr, fragmentsErr, descriptionErr, datesErr)
	if err != nil {
		return domain.Snippet{}, err
	}

	snippet, err := domain.NewSnippet(
		c.ids.NewSnippetID(),
		title,
		description,
		c.folderIDs[record.Folder],
		fragments,
		c.carriedTags(record.Tags),
		stamped.created,
		latestUpdate(stamped.modified, fragments),
	)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("build snippet: %w", err)
	}

	c.recordLosses(record, title)

	return snippet, nil
}

func (r snippetRecord) linkedToGist() bool {
	return r.GistIdentifier != ""
}

func (r snippetRecord) attached() bool {
	return slices.ContainsFunc(r.Fragments, fragmentRecord.attached)
}

func (r snippetRecord) named(err error) error {
	return fmt.Errorf("snippet %s: %w", r.UUID, err)
}

// SnippetsLab leaves a Snippet's date alone on some Fragment changes, such as a
// Language change, so a Fragment can be newer than its Snippet.
func latestUpdate(modified time.Time, fragments []domain.Fragment) time.Time {
	latest := modified

	for _, fragment := range fragments {
		if fragment.UpdatedAt().After(latest) {
			latest = fragment.UpdatedAt()
		}
	}

	return latest
}
