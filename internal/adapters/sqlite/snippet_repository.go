package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
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

func (r *SnippetRepository) Update(ctx context.Context, snippet domain.Snippet, loadedUpdatedAt time.Time) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return updateSnippet(ctx, queries, snippet, loadedUpdatedAt)
	})
	if err != nil {
		return fmt.Errorf("sqlite.SnippetRepository.Update: %w", err)
	}

	return nil
}

func (r *SnippetRepository) Delete(ctx context.Context, id domain.SnippetID) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return deleteSnippet(ctx, queries, id)
	})
	if err != nil {
		return fmt.Errorf("sqlite.SnippetRepository.Delete: %w", err)
	}

	return nil
}

func (r *SnippetRepository) Find(ctx context.Context, id domain.SnippetID) (domain.Snippet, error) {
	snippet, err := r.loadOne(ctx, snippetByID(id))
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("sqlite.SnippetRepository.Find: %w", err)
	}

	return snippet, nil
}

func (r *SnippetRepository) List(ctx context.Context) ([]domain.Snippet, error) {
	snippets, err := r.loadAll(ctx, allSnippets())
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.List: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) ListInFolder(
	ctx context.Context,
	folderID domain.FolderID,
	order domain.SortOrder,
) ([]domain.Snippet, error) {
	snippets, err := r.loadAll(ctx, snippetsInFolder(folderID, order))
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.ListInFolder: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) ListWithTag(
	ctx context.Context,
	tagID domain.TagID,
	order domain.SortOrder,
) ([]domain.Snippet, error) {
	snippets, err := r.loadAll(ctx, snippetsWithTag(tagID, order))
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.ListWithTag: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) CountByFolder(ctx context.Context) (map[domain.FolderID]int, error) {
	rows, err := readRows(ctx, r.db, (*sqlcgen.Queries).CountSnippetsByFolder)
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.CountByFolder: %w", err)
	}

	counts := make(map[domain.FolderID]int, len(rows))
	for _, row := range rows {
		counts[folderFromColumn(row.FolderID)] = int(row.SnippetCount)
	}

	return counts, nil
}

func (r *SnippetRepository) CountByTag(ctx context.Context) (map[domain.TagID]int, error) {
	rows, err := readRows(ctx, r.db, (*sqlcgen.Queries).CountSnippetsByTag)
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.CountByTag: %w", err)
	}

	counts := make(map[domain.TagID]int, len(rows))
	for _, row := range rows {
		counts[entityID[domain.TagID](row.TagID)] = int(row.SnippetCount)
	}

	return counts, nil
}

func updateSnippet(
	ctx context.Context, queries *sqlcgen.Queries, snippet domain.Snippet, loadedUpdatedAt time.Time,
) error {
	updated, err := queries.UpdateSnippet(ctx, updateSnippetParams(snippet, loadedUpdatedAt))
	if err != nil {
		return fmt.Errorf("update snippet: %w", err)
	}

	if updated == 0 {
		return whyNotUpdated(ctx, queries, snippet.ID())
	}

	for _, fragment := range snippet.Fragments() {
		err = queries.UpdateFragment(ctx, updateFragmentParams(fragment))
		if err != nil {
			return fmt.Errorf("update fragment: %w", err)
		}
	}

	return nil
}

func deleteSnippet(ctx context.Context, queries *sqlcgen.Queries, id domain.SnippetID) error {
	deleted, err := queries.DeleteSnippet(ctx, columnID(id))
	if err != nil {
		return fmt.Errorf("delete snippet: %w", err)
	}

	if deleted == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func whyNotUpdated(ctx context.Context, queries *sqlcgen.Queries, id domain.SnippetID) error {
	_, err := queries.GetSnippet(ctx, columnID(id))
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrNotFound
	case err != nil:
		return fmt.Errorf("get snippet: %w", err)
	}

	return domain.ErrConflict
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

	for _, tag := range snippet.Tags() {
		err = queries.InsertSnippetTag(ctx, insertSnippetTagParams(snippet.ID(), tag))
		if err != nil {
			return fmt.Errorf("insert snippet tag: %w", err)
		}
	}

	return nil
}
