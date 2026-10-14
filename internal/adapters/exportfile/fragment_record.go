package exportfile

import (
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

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

func (r fragmentRecord) fragment() (domain.Fragment, error) {
	id, idErr := parseID[domain.FragmentID](r.ID)
	title, titleErr := value.NewFragmentTitle(r.Title)
	language, languageErr := value.NewLanguage(r.Language)
	content, contentErr := value.NewContent(r.Content)
	stamped, datesErr := parseDates(r.CreatedAt, r.UpdatedAt)

	err := errors.Join(idErr, titleErr, languageErr, contentErr, datesErr)
	if err != nil {
		return domain.Fragment{}, r.named(err)
	}

	fragment, err := domain.NewFragment(id, title, language, content, stamped.created, stamped.updated)
	if err != nil {
		return domain.Fragment{}, r.named(err)
	}

	return fragment, nil
}

func (r fragmentRecord) named(err error) error {
	return fmt.Errorf("fragment %s: %w", r.ID, err)
}
