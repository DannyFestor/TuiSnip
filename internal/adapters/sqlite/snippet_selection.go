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

type snippetListing[A any] func(ctx context.Context, arg A) ([]sqlcgen.Snippet, error)

type orderedListings[A any] struct {
	byTitle   snippetListing[A]
	byUpdated snippetListing[A]
	byCreated snippetListing[A]
}

type snippetSelection struct {
	snippets  func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error)
	fragments func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error)
	tags      func(ctx context.Context, queries *sqlcgen.Queries) ([]snippetTagRow, error)
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
		tags: func(ctx context.Context, queries *sqlcgen.Queries) ([]snippetTagRow, error) {
			return snippetTagRowsOf(queries.ListSnippetTagsBySnippet(ctx, column))
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
		tags: func(ctx context.Context, queries *sqlcgen.Queries) ([]snippetTagRow, error) {
			return queries.ListSnippetTags(ctx)
		},
	}
}

func snippetsInFolder(folderID domain.FolderID, order domain.SortOrder) snippetSelection {
	column := folderColumn(folderID)

	return snippetSelection{
		snippets: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error) {
			listings := orderedListings[*sqltype.ID]{
				byTitle:   queries.ListSnippetsInFolderByTitle,
				byUpdated: queries.ListSnippetsInFolderByUpdated,
				byCreated: queries.ListSnippetsInFolderByCreated,
			}

			return listings.run(ctx, order, column)
		},
		fragments: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error) {
			return queries.ListFragmentsInFolder(ctx, column)
		},
		tags: func(ctx context.Context, queries *sqlcgen.Queries) ([]snippetTagRow, error) {
			return snippetTagRowsOf(queries.ListSnippetTagsInFolder(ctx, column))
		},
	}
}

func snippetsWithTag(tagID domain.TagID, order domain.SortOrder) snippetSelection {
	column := columnID(tagID)

	return snippetSelection{
		snippets: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Snippet, error) {
			listings := orderedListings[sqltype.ID]{
				byTitle:   queries.ListSnippetsWithTagByTitle,
				byUpdated: queries.ListSnippetsWithTagByUpdated,
				byCreated: queries.ListSnippetsWithTagByCreated,
			}

			return listings.run(ctx, order, column)
		},
		fragments: func(ctx context.Context, queries *sqlcgen.Queries) ([]sqlcgen.Fragment, error) {
			return queries.ListFragmentsWithTag(ctx, column)
		},
		tags: func(ctx context.Context, queries *sqlcgen.Queries) ([]snippetTagRow, error) {
			return snippetTagRowsOf(queries.ListSnippetTagsWithTag(ctx, column))
		},
	}
}

func (l orderedListings[A]) run(ctx context.Context, order domain.SortOrder, arg A) ([]sqlcgen.Snippet, error) {
	listing, err := l.by(order)
	if err != nil {
		return nil, err
	}

	return listing(ctx, arg)
}

func (l orderedListings[A]) by(order domain.SortOrder) (snippetListing[A], error) {
	switch order {
	case domain.SortOrderTitle:
		return l.byTitle, nil
	case domain.SortOrderUpdated:
		return l.byUpdated, nil
	case domain.SortOrderCreated:
		return l.byCreated, nil
	}

	return nil, fmt.Errorf("list snippets: %w: %q", domain.ErrInvalidSortOrder, order)
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

	tags, err := s.tags(ctx, queries)
	if err != nil {
		return loadedRows{}, fmt.Errorf("select tags: %w", err)
	}

	return loadedRows{snippets: snippets, fragments: fragmentsBySnippet(fragments), tags: tagsBySnippet(tags)}, nil
}
