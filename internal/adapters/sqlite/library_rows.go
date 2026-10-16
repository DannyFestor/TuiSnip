package sqlite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type libraryRows struct {
	folders  []sqlcgen.Folder
	tags     []sqlcgen.Tag
	snippets loadedRows
}

type rowRebuilder[R, E any] struct {
	kind    string
	idAttr  func(row R) slog.Attr
	rebuild func(row R) (E, error)
}

func selectLibraryRows(ctx context.Context, queries *sqlcgen.Queries) (libraryRows, error) {
	folders, err := queries.ListLibraryFolders(ctx)
	if err != nil {
		return libraryRows{}, fmt.Errorf("list folders: %w", err)
	}

	tags, err := queries.ListTags(ctx)
	if err != nil {
		return libraryRows{}, fmt.Errorf("list tags: %w", err)
	}

	snippets, err := librarySnippets().selectRows(ctx, queries)
	if err != nil {
		return libraryRows{}, err
	}

	return libraryRows{folders: folders, tags: tags, snippets: snippets}, nil
}

func (r libraryRows) rebuild(ctx context.Context, logger *slog.Logger) (domain.Library, error) {
	folders, foldersErr := rebuildEach(ctx, logger, r.folders, folderRebuilder())
	tags, tagsErr := rebuildEach(ctx, logger, r.tags, tagRebuilder())
	snippets, snippetsErr := rebuildEach(ctx, logger, r.snippets.snippets, r.snippetRebuilder())

	err := errors.Join(foldersErr, tagsErr, snippetsErr)
	if err != nil {
		return domain.Library{}, err
	}

	return domain.NewLibrary(folders, tags, snippets), nil
}

func (r libraryRows) snippetRebuilder() rowRebuilder[sqlcgen.Snippet, domain.LibraryEntry[domain.Snippet]] {
	return rowRebuilder[sqlcgen.Snippet, domain.LibraryEntry[domain.Snippet]]{
		kind: "snippet",
		idAttr: func(row sqlcgen.Snippet) slog.Attr {
			return slog.String(keySnippetID, entityID[domain.SnippetID](row.ID).String())
		},
		rebuild: func(row sqlcgen.Snippet) (domain.LibraryEntry[domain.Snippet], error) {
			snippet, err := snippetFromRows(row, r.snippets.fragments[row.ID], r.snippets.tags[row.ID])
			if err != nil {
				return domain.LibraryEntry[domain.Snippet]{}, err
			}

			return entryFromColumns(snippet, row.DeletedAt, row.FormerFolderPath)
		},
	}
}

func folderRebuilder() rowRebuilder[sqlcgen.Folder, domain.LibraryEntry[domain.Folder]] {
	return rowRebuilder[sqlcgen.Folder, domain.LibraryEntry[domain.Folder]]{
		kind: "folder",
		idAttr: func(row sqlcgen.Folder) slog.Attr {
			return slog.String(keyFolderID, entityID[domain.FolderID](row.ID).String())
		},
		rebuild: func(row sqlcgen.Folder) (domain.LibraryEntry[domain.Folder], error) {
			folder, err := folderFromRow(row)
			if err != nil {
				return domain.LibraryEntry[domain.Folder]{}, err
			}

			return entryFromColumns(folder, row.DeletedAt, row.FormerParentPath)
		},
	}
}

func tagRebuilder() rowRebuilder[sqlcgen.Tag, domain.Tag] {
	return rowRebuilder[sqlcgen.Tag, domain.Tag]{
		kind: "tag",
		idAttr: func(row sqlcgen.Tag) slog.Attr {
			return slog.String(keyTagID, entityID[domain.TagID](row.ID).String())
		},
		rebuild: standaloneTagFromRow,
	}
}

func rebuildEach[R, E any](
	ctx context.Context, logger *slog.Logger, rows []R, rebuilder rowRebuilder[R, E],
) ([]E, error) {
	rebuilt := make([]E, 0, len(rows))

	var errs []error

	for _, row := range rows {
		entity, err := rebuilder.rebuild(row)
		if err != nil {
			errs = append(errs, rebuilder.reported(ctx, logger, row, err))

			continue
		}

		rebuilt = append(rebuilt, entity)
	}

	return rebuilt, errors.Join(errs...)
}

func (b rowRebuilder[R, E]) reported(ctx context.Context, logger *slog.Logger, row R, err error) error {
	id := b.idAttr(row)
	logger.ErrorContext(ctx, "library row is corrupt", id, slog.String(keyError, err.Error()))

	return fmt.Errorf("%s %s: %w", b.kind, id.Value, err)
}
