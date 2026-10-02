package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSnippetRepository(database *Database, logger *slog.Logger) *SnippetRepository {
	return &SnippetRepository{db: database.db, logger: logger}
}

func (r *SnippetRepository) Insert(ctx context.Context, snippet domain.Snippet) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return insertSnippet(ctx, queries, snippet)
	})
	if err != nil {
		return fmt.Errorf("sqlite.SnippetRepository.Insert: %w", err)
	}

	return nil
}

func (r *SnippetRepository) Find(ctx context.Context, id domain.SnippetID) (domain.Snippet, error) {
	var snippet domain.Snippet

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var findErr error

		snippet, findErr = r.findWith(ctx, queries, id)

		return findErr
	})
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("sqlite.SnippetRepository.Find: %w", err)
	}

	return snippet, nil
}

func (r *SnippetRepository) List(ctx context.Context) ([]domain.Snippet, error) {
	var snippets []domain.Snippet

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var listErr error

		snippets, listErr = r.listWith(ctx, queries)

		return listErr
	})
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.List: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) ListInFolder(ctx context.Context, folderID domain.FolderID) ([]domain.Snippet, error) {
	var snippets []domain.Snippet

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var listErr error

		snippets, listErr = r.listInFolderWith(ctx, queries, folderColumn(folderID))

		return listErr
	})
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.ListInFolder: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) findWith(
	ctx context.Context, queries *sqlcgen.Queries, id domain.SnippetID,
) (domain.Snippet, error) {
	row, err := queries.GetSnippet(ctx, columnID(id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Snippet{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Snippet{}, fmt.Errorf("get snippet: %w", err)
	}

	fragments, err := queries.ListFragmentsBySnippet(ctx, row.ID)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("list fragments of snippet: %w", err)
	}

	return r.rebuild(ctx, row, fragments)
}

func (r *SnippetRepository) listWith(ctx context.Context, queries *sqlcgen.Queries) ([]domain.Snippet, error) {
	rows, err := queries.ListSnippets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}

	fragmentRows, err := queries.ListFragments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list fragments: %w", err)
	}

	return r.rebuildAll(ctx, rows, fragmentRows)
}

func (r *SnippetRepository) listInFolderWith(
	ctx context.Context, queries *sqlcgen.Queries, folderID *sqltype.ID,
) ([]domain.Snippet, error) {
	rows, err := queries.ListSnippetsInFolder(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("list snippets in folder: %w", err)
	}

	fragmentRows, err := queries.ListFragmentsInFolder(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("list fragments in folder: %w", err)
	}

	return r.rebuildAll(ctx, rows, fragmentRows)
}

func (r *SnippetRepository) rebuildAll(
	ctx context.Context, rows []sqlcgen.Snippet, fragmentRows []sqlcgen.Fragment,
) ([]domain.Snippet, error) {
	fragments := fragmentsBySnippet(fragmentRows)
	snippets := make([]domain.Snippet, 0, len(rows))

	for _, row := range rows {
		snippet, err := r.rebuild(ctx, row, fragments[row.ID])
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, snippet)
	}

	return snippets, nil
}

func (r *SnippetRepository) rebuild(
	ctx context.Context, row sqlcgen.Snippet, fragments []sqlcgen.Fragment,
) (domain.Snippet, error) {
	snippet, err := snippetFromRows(row, fragments)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"snippet row is corrupt",
			slog.String(keySnippetID, entityID[domain.SnippetID](row.ID).String()),
		)
	}

	return snippet, err
}

func insertSnippet(ctx context.Context, queries *sqlcgen.Queries, snippet domain.Snippet) error {
	err := queries.InsertSnippet(ctx, insertSnippetParams(snippet))
	if err != nil {
		return fmt.Errorf("insert snippet: %w", err)
	}

	for position, fragment := range snippet.Fragments() {
		err = queries.InsertFragment(ctx, insertFragmentParams(snippet.ID(), position, fragment))
		if err != nil {
			return fmt.Errorf("insert fragment: %w", err)
		}
	}

	return nil
}
