package sqlite

import (
	"context"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type loadedRows struct {
	snippets  []sqlcgen.Snippet
	fragments map[sqltype.ID][]sqlcgen.Fragment
}

func (r *SnippetRepository) loadOne(ctx context.Context, selection snippetSelection) (domain.Snippet, error) {
	rows, err := r.read(ctx, selection)
	if err != nil {
		return domain.Snippet{}, err
	}

	if len(rows.snippets) == 0 {
		return domain.Snippet{}, domain.ErrNotFound
	}

	return r.rebuild(ctx, rows, rows.snippets[0], slog.LevelError)
}

func (r *SnippetRepository) loadAll(ctx context.Context, selection snippetSelection) ([]domain.Snippet, error) {
	rows, err := r.read(ctx, selection)
	if err != nil {
		return nil, err
	}

	snippets := make([]domain.Snippet, 0, len(rows.snippets))

	for _, row := range rows.snippets {
		snippet, rebuildErr := r.rebuild(ctx, rows, row, slog.LevelWarn)
		if rebuildErr != nil {
			continue
		}

		snippets = append(snippets, snippet)
	}

	return snippets, nil
}

func (r *SnippetRepository) read(ctx context.Context, selection snippetSelection) (loadedRows, error) {
	var rows loadedRows

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var selectErr error

		rows, selectErr = selection.selectRows(ctx, queries)

		return selectErr
	})

	return rows, err
}

func (r *SnippetRepository) rebuild(
	ctx context.Context, rows loadedRows, row sqlcgen.Snippet, corruptLevel slog.Level,
) (domain.Snippet, error) {
	snippet, err := snippetFromRows(row, rows.fragments[row.ID])
	if err != nil {
		r.logger.LogAttrs(
			ctx,
			corruptLevel,
			"snippet row is corrupt",
			slog.String(keySnippetID, entityID[domain.SnippetID](row.ID).String()),
			slog.String(keyError, err.Error()),
		)
	}

	return snippet, err
}
