package browse

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderTree struct {
	folders FolderLister
	counter SnippetCounter
	trash   TrashItemCounter
}

func NewFolderTree(folders FolderLister, counter SnippetCounter, trash TrashItemCounter) (*FolderTree, error) {
	err := errors.Join(
		domain.RequireDependency("folders", folders),
		domain.RequireDependency("counter", counter),
		domain.RequireDependency("trash", trash),
	)
	if err != nil {
		return nil, fmt.Errorf("browse.NewFolderTree: %w", err)
	}

	return &FolderTree{folders: folders, counter: counter, trash: trash}, nil
}

func (f *FolderTree) Run(ctx context.Context, _ FolderTreeInput) (Tree, error) {
	folders, err := f.folders.List(ctx)
	if err != nil {
		return Tree{}, fmt.Errorf("browse.FolderTree: %w", err)
	}

	counts, err := f.counter.CountByFolder(ctx)
	if err != nil {
		return Tree{}, fmt.Errorf("browse.FolderTree: %w", err)
	}

	trashCount, err := f.trash.CountItems(ctx)
	if err != nil {
		return Tree{}, fmt.Errorf("browse.FolderTree: %w", err)
	}

	root := domain.FolderID{}

	return Tree{
		RootSnippetCount: counts[root],
		Folders:          nodesUnder(root, childrenByParent(folders), counts),
		TrashCount:       trashCount,
	}, nil
}

func childrenByParent(folders []domain.Folder) map[domain.FolderID][]domain.Folder {
	children := make(map[domain.FolderID][]domain.Folder, len(folders))
	for _, folder := range folders {
		children[folder.ParentID()] = append(children[folder.ParentID()], folder)
	}

	for _, siblings := range children {
		slices.SortFunc(siblings, domain.CompareFolders)
	}

	return children
}

func nodesUnder(
	parent domain.FolderID, children map[domain.FolderID][]domain.Folder, counts map[domain.FolderID]int,
) []FolderNode {
	siblings := children[parent]
	if len(siblings) == 0 {
		return nil
	}

	nodes := make([]FolderNode, 0, len(siblings))
	for _, folder := range siblings {
		nodes = append(nodes, FolderNode{
			Folder:       folder,
			SnippetCount: counts[folder.ID()],
			Children:     nodesUnder(folder.ID(), children, counts),
		})
	}

	return nodes
}
