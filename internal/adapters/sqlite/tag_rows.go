package sqlite

import (
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type snippetTagRow = sqlcgen.ListSnippetTagsRow

type snippetTagRowShape interface {
	~struct {
		SnippetID sqltype.ID
		Tag       sqlcgen.Tag
	}
}

func snippetTagRowsOf[R snippetTagRowShape](rows []R, err error) ([]snippetTagRow, error) {
	if err != nil {
		return nil, err
	}

	converted := make([]snippetTagRow, 0, len(rows))
	for _, row := range rows {
		converted = append(converted, snippetTagRow(row))
	}

	return converted, nil
}

func tagsBySnippet(rows []snippetTagRow) map[sqltype.ID][]sqlcgen.Tag {
	grouped := make(map[sqltype.ID][]sqlcgen.Tag, len(rows))
	for _, row := range rows {
		grouped[row.SnippetID] = append(grouped[row.SnippetID], row.Tag)
	}

	return grouped
}

func tagsFromRows(rows []sqlcgen.Tag) ([]domain.Tag, error) {
	tags := make([]domain.Tag, 0, len(rows))
	for _, row := range rows {
		tag, err := tagFromRow(row)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func tagFromRow(row sqlcgen.Tag) (domain.Tag, error) {
	name, err := value.NewTagName(row.Name)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("rebuild tag name: %w", err)
	}

	tag, err := domain.NewTag(
		entityID[domain.TagID](row.ID),
		name,
		time.Time(row.CreatedAt),
		time.Time(row.UpdatedAt),
	)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("rebuild tag: %w", err)
	}

	return tag, nil
}
