package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

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
