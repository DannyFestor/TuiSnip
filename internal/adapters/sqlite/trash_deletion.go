package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type rowCounts struct {
	folders  int64
	snippets int64
}

func deleteSnippetTrashItem(ctx context.Context, queries *sqlcgen.Queries, id domain.SnippetID) error {
	deleted, err := queries.DeleteSnippetTrashItem(ctx, columnID(id))
	if err != nil {
		return fmt.Errorf("delete snippet trash item: %w", err)
	}

	if deleted == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func deleteFolderTrashItem(ctx context.Context, queries *sqlcgen.Queries, id domain.FolderID) error {
	tree, err := loadFolderTree(ctx, queries)
	if err != nil {
		return err
	}

	if !tree.isTrashItem(id) {
		return domain.ErrNotFound
	}

	err = detachTrashItemsIn(ctx, queries, tree, id)
	if err != nil {
		return err
	}

	err = queries.DeleteFolderTrashItem(ctx, columnID(id))
	if err != nil {
		return fmt.Errorf("delete folder trash item: %w", err)
	}

	return nil
}

func detachTrashItemsIn(ctx context.Context, queries *sqlcgen.Queries, tree folderTree, id domain.FolderID) error {
	reached := tree.cascadeFrom(id)

	for _, folder := range reached.trashFolders {
		err := detachFrom(tree, folder.ParentID(), func(formerPath string) error {
			return queries.DetachFolderTrashItem(
				ctx, sqlcgen.DetachFolderTrashItemParams{FormerParentPath: formerPath, ID: columnID(folder.ID())},
			)
		})
		if err != nil {
			return err
		}
	}

	return detachSnippetTrashItemsIn(ctx, queries, tree, reached)
}

func detachSnippetTrashItemsIn(
	ctx context.Context, queries *sqlcgen.Queries, tree folderTree, reached cascade,
) error {
	filed, err := queries.ListFiledSnippetTrashItems(ctx)
	if err != nil {
		return fmt.Errorf("list filed snippet trash items: %w", err)
	}

	for _, snippet := range filed {
		folderID := folderFromColumn(snippet.FolderID)
		if !reached.folders[folderID] {
			continue
		}

		err = detachFrom(tree, folderID, func(formerPath string) error {
			return queries.DetachSnippetTrashItem(
				ctx, sqlcgen.DetachSnippetTrashItemParams{FormerFolderPath: formerPath, ID: snippet.ID},
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func detachFrom(tree folderTree, folderID domain.FolderID, detach func(formerPath string) error) error {
	formerPath, err := formerPathColumn(tree.paths.To(folderID))
	if err != nil {
		return err
	}

	err = detach(formerPath)
	if err != nil {
		return fmt.Errorf("detach trash item: %w", err)
	}

	return nil
}

func emptyTrash(ctx context.Context, queries *sqlcgen.Queries, items domain.TrashItemIDs) error {
	for _, id := range items.Folders {
		err := skippingGone(deleteFolderTrashItem(ctx, queries, id))
		if err != nil {
			return err
		}
	}

	for _, id := range items.Snippets {
		err := skippingGone(deleteSnippetTrashItem(ctx, queries, id))
		if err != nil {
			return err
		}
	}

	return nil
}

func skippingGone(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}

	return err
}

func countRows(ctx context.Context, queries *sqlcgen.Queries) (rowCounts, error) {
	folders, err := queries.CountFolders(ctx)
	if err != nil {
		return rowCounts{}, fmt.Errorf("count folders: %w", err)
	}

	snippets, err := queries.CountSnippets(ctx)
	if err != nil {
		return rowCounts{}, fmt.Errorf("count snippets: %w", err)
	}

	return rowCounts{folders: folders, snippets: snippets}, nil
}

func (c rowCounts) emptiedTo(after rowCounts) domain.TrashEmptied {
	return domain.TrashEmptied{Folders: int(c.folders - after.folders), Snippets: int(c.snippets - after.snippets)}
}
