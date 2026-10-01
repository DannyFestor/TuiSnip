package sqlite

import (
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func snippetFromRows(row sqlcgen.Snippet, fragmentRows []sqlcgen.Fragment) (domain.Snippet, error) {
	fragments, fragmentsErr := fragmentsFromRows(fragmentRows)
	title, titleErr := value.NewTitle(row.Title)
	description, descriptionErr := value.NewDescription(row.Description)

	err := errors.Join(fragmentsErr, titleErr, descriptionErr)
	if err != nil {
		return domain.Snippet{}, corrupt(err)
	}

	snippet, err := domain.NewSnippet(
		entityID[domain.SnippetID](row.ID),
		title,
		description,
		folderFromColumn(row.FolderID),
		fragments,
		time.Time(row.CreatedAt),
		time.Time(row.UpdatedAt),
	)
	if err != nil {
		return domain.Snippet{}, corrupt(err)
	}

	return snippet, nil
}

func fragmentsBySnippet(rows []sqlcgen.Fragment) map[sqltype.ID][]sqlcgen.Fragment {
	grouped := make(map[sqltype.ID][]sqlcgen.Fragment, len(rows))
	for _, row := range rows {
		grouped[row.SnippetID] = append(grouped[row.SnippetID], row)
	}

	return grouped
}

func fragmentsFromRows(rows []sqlcgen.Fragment) ([]domain.Fragment, error) {
	fragments := make([]domain.Fragment, 0, len(rows))
	for _, row := range rows {
		fragment, err := fragmentFromRow(row)
		if err != nil {
			return nil, err
		}

		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func fragmentFromRow(row sqlcgen.Fragment) (domain.Fragment, error) {
	language, languageErr := value.NewLanguage(row.Language)
	content, contentErr := value.NewContent(row.Content)

	err := errors.Join(languageErr, contentErr)
	if err != nil {
		return domain.Fragment{}, err
	}

	fragment, err := domain.NewFragment(
		entityID[domain.FragmentID](row.ID),
		language,
		content,
		time.Time(row.CreatedAt),
		time.Time(row.UpdatedAt),
	)
	if err != nil {
		return domain.Fragment{}, fmt.Errorf("rebuild fragment: %w", err)
	}

	return fragment, nil
}

func corrupt(err error) error {
	return fmt.Errorf("%w: %w", domain.ErrCorruptRecord, err)
}
