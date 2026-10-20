package sqlite

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type folderTree struct {
	children map[domain.FolderID][]domain.Folder
	trash    domain.Trash
	paths    domain.FolderPaths
}

type cascade struct {
	folders      map[domain.FolderID]bool
	trashFolders []domain.Folder
}

func loadFolderTree(ctx context.Context, queries *sqlcgen.Queries) (folderTree, error) {
	rows, err := queries.ListLibraryFolders(ctx)
	if err != nil {
		return folderTree{}, fmt.Errorf("list folders: %w", err)
	}

	entries := make([]domain.LibraryEntry[domain.Folder], 0, len(rows))

	for index := range rows {
		entry, rebuildErr := folderRebuilder().rebuild(rows[index])
		if rebuildErr != nil {
			return folderTree{}, rebuildErr
		}

		entries = append(entries, entry)
	}

	return folderTreeOf(domain.NewLibrary(entries, nil, nil)), nil
}

func folderTreeOf(folders domain.Library) folderTree {
	children := make(map[domain.FolderID][]domain.Folder, len(folders.Folders))
	for _, folder := range folders.Folders {
		children[folder.ParentID()] = append(children[folder.ParentID()], folder)
	}

	return folderTree{
		children: children,
		trash:    folders.Trash,
		paths:    domain.NewFolderPaths(folders.Folders, folders.Trash),
	}
}

func (t folderTree) isTrashItem(id domain.FolderID) bool {
	_, trashed := t.trash.Folders[id]

	return trashed
}

func (t folderTree) cascadeFrom(id domain.FolderID) cascade {
	reached := cascade{folders: map[domain.FolderID]bool{id: true}, trashFolders: nil}
	pending := []domain.FolderID{id}

	for len(pending) > 0 {
		parentID := pending[0]
		pending = pending[1:]

		for _, child := range t.children[parentID] {
			switch {
			case t.isTrashItem(child.ID()):
				reached.trashFolders = append(reached.trashFolders, child)
			case !reached.folders[child.ID()]:
				reached.folders[child.ID()] = true
				pending = append(pending, child.ID())
			}
		}
	}

	return reached
}
