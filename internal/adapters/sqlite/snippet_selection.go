package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type folderListing func(ctx context.Context, folderID *sqltype.ID) ([]sqlcgen.Snippet, error)

type snippetSelection struct {
	snippets  func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error)
	fragments func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error)
}

func snippetByID(id domain.SnippetID) snippetSelection {
	column := columnID(id)

	return snippetSelection{
		snippets: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error) {
			row, err := queries.GetSnippet(ctx, column)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}

			if err != nil {
				return nil, fmt.Errorf("get snippet: %w", err)
			}

			return []sqlcgen.Snippet{row}, nil
		},
		fragments: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error) {
			return queries.ListFragmentsBySnippet(ctx, column)
		},
	}
}

func allSnippets() snippetSelection {
	return snippetSelection{
		snippets: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error) {
			return queries.ListSnippets(ctx)
		},
		fragments: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error) {
			return queries.ListFragments(ctx)
		},
	}
}

func snippetsInFolder(folderID domain.FolderID, order domain.SortOrder) snippetSelection {
	column := folderColumn(folderID)

	return snippetSelection{
		snippets: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error) {
			list, err := folderListingBy(queries, order)
			if err != nil {
				return nil, err
			}

			return list(ctx, column)
		},
		fragments: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error) {
			return queries.ListFragmentsInFolder(ctx, column)
		},
	}
}

func folderListingBy(queries *sqlcgen.Queries, order domain.SortOrder) (folderListing, error) {
	switch order {
	case domain.SortOrderTitle:
		return queries.ListSnippetsInFolderByTitle, nil
	case domain.SortOrderUpdated:
		return queries.ListSnippetsInFolderByUpdated, nil
	case domain.SortOrderCreated:
		return queries.ListSnippetsInFolderByCreated, nil
	}

	return nil, fmt.Errorf("list snippets in folder: %w: %q", domain.ErrInvalidSortOrder, order)
}

func (s snippetSelection) selectRows(ctx context.Context, queries *sqlcgen.Queries) (loadedRows, error) {
	snippets, err := s.snippets(ctx, queries)
	if err != nil {
		return loadedRows{}, fmt.Errorf("select snippets: %w", err)
	}

	fragments, err := s.fragments(ctx, queries)
	if err != nil {
		return loadedRows{}, fmt.Errorf("select fragments: %w", err)
	}

	return loadedRows{snippets: snippets, fragments: fragmentsBySnippet(fragments)}, nil
}
