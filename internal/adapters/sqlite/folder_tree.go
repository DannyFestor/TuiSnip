package sqlite

import (
	"context"
	"fmt"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type folderNode struct {
	id               domain.FolderID
	parentID         domain.FolderID
	name             string
	formerParentPath string
	trashed          bool
}

type folderTree struct {
	nodes    map[domain.FolderID]folderNode
	children map[domain.FolderID][]folderNode
}

type cascade struct {
	folders      map[domain.FolderID]bool
	trashFolders []folderNode
}

func loadFolderTree(ctx context.Context, queries *sqlcgen.Queries) (folderTree, error) {
	rows, err := queries.ListLibraryFolders(ctx)
	if err != nil {
		return folderTree{}, fmt.Errorf("list folders: %w", err)
	}

	tree := folderTree{
		nodes:    make(map[domain.FolderID]folderNode, len(rows)),
		children: make(map[domain.FolderID][]folderNode, len(rows)),
	}

	for index := range rows {
		node := folderNodeOf(rows[index])
		tree.nodes[node.id] = node
		tree.children[node.parentID] = append(tree.children[node.parentID], node)
	}

	return tree, nil
}

func folderNodeOf(row sqlcgen.Folder) folderNode {
	return folderNode{
		id:               entityID[domain.FolderID](row.ID),
		parentID:         folderFromColumn(row.ParentID),
		name:             row.Name,
		formerParentPath: row.FormerParentPath,
		trashed:          row.DeletedAt != nil,
	}
}

func (t folderTree) isTrashItem(id domain.FolderID) bool {
	return t.nodes[id].trashed
}

func (t folderTree) cascadeFrom(id domain.FolderID) cascade {
	reached := cascade{folders: map[domain.FolderID]bool{id: true}, trashFolders: nil}
	pending := []domain.FolderID{id}

	for len(pending) > 0 {
		parentID := pending[0]
		pending = pending[1:]

		for _, child := range t.children[parentID] {
			switch {
			case child.trashed:
				reached.trashFolders = append(reached.trashFolders, child)
			case !reached.folders[child.id]:
				reached.folders[child.id] = true
				pending = append(pending, child.id)
			}
		}
	}

	return reached
}

func (t folderTree) placeOf(id domain.FolderID) ([]string, error) {
	lineage := t.lineageOf(id)
	if len(lineage) == 0 {
		return []string{}, nil
	}

	formerPath, err := formerPathFromColumn(lineage[len(lineage)-1].formerParentPath)
	if err != nil {
		return nil, corrupt(err)
	}

	names := formerPath.Strings()
	for _, node := range slices.Backward(lineage) {
		names = append(names, node.name)
	}

	return names, nil
}

func (t folderTree) lineageOf(id domain.FolderID) []folderNode {
	var lineage []folderNode

	visited := make(map[domain.FolderID]bool)

	for node, known := t.nodes[id]; known && !visited[id]; node, known = t.nodes[id] {
		visited[id] = true

		lineage = append(lineage, node)
		id = node.parentID
	}

	return lineage
}
