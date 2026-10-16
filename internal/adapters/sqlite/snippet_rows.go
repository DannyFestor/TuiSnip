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

func snippetFromRows(
	row sqlcgen.Snippet, fragmentRows []sqlcgen.Fragment, tagRows []sqlcgen.Tag,
) (domain.Snippet, error) {
	fragments, fragmentsErr := fragmentsFromRows(fragmentRows)
	tags, tagsErr := tagsFromRows(tagRows)
	title, titleErr := value.NewTitle(row.Title)
	description, descriptionErr := value.NewDescription(row.Description)

	err := errors.Join(fragmentsErr, tagsErr, titleErr, descriptionErr)
	if err != nil {
		return domain.Snippet{}, corrupt(err)
	}

	snippet, err := domain.NewSnippet(
		entityID[domain.SnippetID](row.ID),
		title,
		description,
		folderFromColumn(row.FolderID),
		fragments,
		tags,
		time.Time(row.CreatedAt),
		time.Time(row.UpdatedAt),
	)
	if err != nil {
		return domain.Snippet{}, corrupt(err)
	}

	return snippet, nil
}

func snippetRowsOf(rows []sqlcgen.LiveSnippet, err error) ([]sqlcgen.Snippet, error) {
	if err != nil {
		return nil, err
	}

	converted := make([]sqlcgen.Snippet, 0, len(rows))
	for index := range rows {
		converted = append(converted, sqlcgen.Snippet(rows[index]))
	}

	return converted, nil
}

func fragmentsBySnippet(rows []sqlcgen.Fragment) map[sqltype.ID][]sqlcgen.Fragment {
	grouped := make(map[sqltype.ID][]sqlcgen.Fragment, len(rows))
	for index := range rows {
		snippetID := rows[index].SnippetID
		grouped[snippetID] = append(grouped[snippetID], rows[index])
	}

	return grouped
}

func fragmentsFromRows(rows []sqlcgen.Fragment) ([]domain.Fragment, error) {
	fragments := make([]domain.Fragment, 0, len(rows))
	for index := range rows {
		fragment, err := fragmentFromRow(rows[index])
		if err != nil {
			return nil, err
		}

		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

func fragmentFromRow(row sqlcgen.Fragment) (domain.Fragment, error) {
	title, titleErr := value.NewFragmentTitle(row.Title)
	language, languageErr := value.NewLanguage(row.Language)
	content, contentErr := value.NewContent(row.Content)

	err := errors.Join(titleErr, languageErr, contentErr)
	if err != nil {
		return domain.Fragment{}, err
	}

	fragment, err := domain.NewFragment(
		entityID[domain.FragmentID](row.ID),
		title,
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
