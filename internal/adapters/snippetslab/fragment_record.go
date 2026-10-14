package snippetslab

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const defaultFragmentTitle = "Fragment"

type fragmentRecord struct {
	Title          string            `json:"title"`
	Language       string            `json:"language"`
	Content        string            `json:"content"`
	Note           string            `json:"note"`
	NoteAttributes []noteAttribute   `json:"noteAttributes"`
	Attachments    []json.RawMessage `json:"attachments"`
	DateCreated    string            `json:"dateCreated"`
	DateModified   string            `json:"dateModified"`
}

func (r fragmentRecord) attached() bool {
	return len(r.Attachments) > 0
}

func (r fragmentRecord) keptTitle() string {
	if r.Title == defaultFragmentTitle {
		return ""
	}

	return r.Title
}

func (r fragmentRecord) label() string {
	title, err := value.NewFragmentTitle(r.keptTitle())
	if err == nil && title.IsSet() {
		return title.String()
	}

	language, _ := languageOfLexer(r.Language)

	return language.String()
}

func (c *conversion) fragments(owner value.Title, records []fragmentRecord) ([]domain.Fragment, error) {
	fragments := make([]domain.Fragment, 0, len(records))

	var errs []error

	for index := range records {
		place := fragmentPlace{snippet: owner, position: index + 1}

		fragment, err := c.fragment(place, records[index])
		if err != nil {
			errs = append(errs, fmt.Errorf("fragment %d: %w", place.position, err))

			continue
		}

		fragments = append(fragments, fragment)
	}

	return fragments, errors.Join(errs...)
}

func (c *conversion) fragment(place fragmentPlace, record fragmentRecord) (domain.Fragment, error) {
	title, titleErr := value.NewFragmentTitle(record.keptTitle())
	language, known := languageOfLexer(record.Language)
	content, contentErr := value.NewContent(record.Content)
	stamped, datesErr := parseDates(record.DateCreated, record.DateModified)

	err := errors.Join(titleErr, contentErr, datesErr)
	if err != nil {
		return domain.Fragment{}, err
	}

	fragment, err := domain.NewFragment(
		c.ids.NewFragmentID(),
		title,
		language,
		content,
		stamped.created,
		stamped.modified,
	)
	if err != nil {
		return domain.Fragment{}, fmt.Errorf("build fragment: %w", err)
	}

	if !known {
		c.losses.FragmentFallbacks = append(c.losses.FragmentFallbacks, place.fellBack(title, record.Language))
	}

	return fragment, nil
}
