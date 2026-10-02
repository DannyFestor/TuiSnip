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
	loader snippetLoader
}

func NewSnippetRepository(database *Database, logger *slog.Logger) *SnippetRepository {
	return &SnippetRepository{
		db:     database.db,
		loader: snippetLoader{db: database.db, logger: logger},
	}
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
	snippet, err := r.loader.loadOne(ctx, snippetByID(id))
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("sqlite.SnippetRepository.Find: %w", err)
	}

	return snippet, nil
}

func (r *SnippetRepository) List(ctx context.Context) ([]domain.Snippet, error) {
	snippets, err := r.loader.loadAll(ctx, allSnippets())
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.List: %w", err)
	}

	return snippets, nil
}

func (r *SnippetRepository) ListInFolder(ctx context.Context, folderID domain.FolderID) ([]domain.Snippet, error) {
	snippets, err := r.loader.loadAll(ctx, snippetsInFolder(folderID))
	if err != nil {
		return nil, fmt.Errorf("sqlite.SnippetRepository.ListInFolder: %w", err)
	}

	return snippets, nil
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
