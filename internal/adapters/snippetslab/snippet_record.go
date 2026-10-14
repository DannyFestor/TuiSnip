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
	UUID         string           `json:"uuid"`
	Title        string           `json:"title"`
	Folder       string           `json:"folder"`
	Tags         []string         `json:"tags"`
	Pinned       bool             `json:"pinned"`
	DateCreated  string           `json:"dateCreated"`
	DateModified string           `json:"dateModified"`
	DateDeleted  string           `json:"dateDeleted"`
	Fragments    []fragmentRecord `json:"fragments"`
}

func (c *conversion) snippet(record snippetRecord) (domain.Snippet, error) {
	title, titleErr := value.NewTitle(record.Title)
	fragments, fragmentsErr := c.fragments(title, record.Fragments)
	description, descriptionErr := descriptionOf(record.Fragments)
	stamped, datesErr := parseDates(record.DateCreated, record.DateModified)

	err := errors.Join(titleErr, fragmentsErr, descriptionErr, datesErr)
	if err != nil {
		return domain.Snippet{}, record.named(err)
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
		return domain.Snippet{}, record.named(err)
	}

	c.recordLostPin(record, title)

	return snippet, nil
}

func liveSnippets(records []snippetRecord) []snippetRecord {
	return slices.DeleteFunc(slices.Clone(records), snippetRecord.trashed)
}

func (r snippetRecord) trashed() bool {
	return r.DateDeleted != ""
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
