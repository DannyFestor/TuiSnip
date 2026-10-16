package domain

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"
)

type Library struct {
	Folders  []Folder
	Tags     []Tag
	Snippets []Snippet
	Trash    Trash
}

func (l Library) InCreatedOrder() Library {
	return Library{
		Folders:  slices.SortedFunc(slices.Values(l.Folders), byCreated(Folder.CreatedAt, Folder.ID)),
		Tags:     slices.SortedFunc(slices.Values(l.Tags), byCreated(Tag.CreatedAt, Tag.ID)),
		Snippets: slices.SortedFunc(slices.Values(l.Snippets), byCreated(Snippet.CreatedAt, Snippet.ID)),
		Trash:    l.Trash,
	}
}

func (l Library) InTrash() Library {
	parentOf := l.parentOf()
	trashedFolders := make(map[FolderID]bool, len(l.Folders))

	for _, folder := range l.Folders {
		trashedFolders[folder.ID()] = l.Trash.holdsFolderOrAncestor(folder.ID(), parentOf)
	}

	return Library{
		Folders: slices.DeleteFunc(slices.Clone(l.Folders), func(folder Folder) bool {
			return !trashedFolders[folder.ID()]
		}),
		Tags: nil,
		Snippets: slices.DeleteFunc(slices.Clone(l.Snippets), func(snippet Snippet) bool {
			_, trashed := l.Trash.Snippets[snippet.ID()]

			return !trashed && !trashedFolders[snippet.FolderID()]
		}),
		Trash: l.Trash,
	}
}

func (l Library) FoldersParentsFirst() ([]Folder, error) {
	depths, err := l.folderDepths()
	if err != nil {
		return nil, fmt.Errorf("domain.Library.FoldersParentsFirst: %w", err)
	}

	return slices.SortedStableFunc(slices.Values(l.Folders), func(a, b Folder) int {
		return cmp.Compare(depths[a.ID()], depths[b.ID()])
	}), nil
}

func (l Library) RequireIntact() error {
	_, foldersErr := l.folderDepths()

	err := errors.Join(foldersErr, l.requireDistinctTagNames(), l.requireSnippetReferences())
	if err != nil {
		return fmt.Errorf("domain.Library.RequireIntact: %w", err)
	}

	return nil
}

func (l Library) ReusingTags(stored []Tag) Library {
	byName := newTagsByName(stored)

	snippets := make([]Snippet, 0, len(l.Snippets))
	for i := range l.Snippets {
		snippets = append(snippets, l.Snippets[i].Retagged(byName.reused(l.Snippets[i].Tags())))
	}

	return Library{Folders: l.Folders, Tags: byName.unmatched(l.Tags), Snippets: snippets, Trash: l.Trash}
}

func (l Library) parentOf() map[FolderID]FolderID {
	parentOf := make(map[FolderID]FolderID, len(l.Folders))
	for _, folder := range l.Folders {
		parentOf[folder.ID()] = folder.ParentID()
	}

	return parentOf
}

func (l Library) folderDepths() (map[FolderID]int, error) {
	parentOf := l.parentOf()
	depths := make(map[FolderID]int, len(l.Folders))

	var errs []error

	for _, folder := range l.Folders {
		depth, err := depthOf(folder, parentOf)
		if err != nil {
			errs = append(errs, fmt.Errorf("folder %s: %w", folder.ID(), err))

			continue
		}

		depths[folder.ID()] = depth
	}

	return depths, errors.Join(errs...)
}

func (l Library) requireDistinctTagNames() error {
	named := make(map[string]bool, len(l.Tags))

	var errs []error

	for _, tag := range l.Tags {
		if named[tag.Name().Key()] {
			errs = append(errs, fmt.Errorf("tag %s: %w", tag.ID(), ErrTagNameTaken))
		}

		named[tag.Name().Key()] = true
	}

	return errors.Join(errs...)
}

func (l Library) requireSnippetReferences() error {
	folderIDs := idSet(l.Folders, Folder.ID)
	tagIDs := idSet(l.Tags, Tag.ID)

	errs := make([]error, 0, len(l.Snippets))
	for i := range l.Snippets {
		errs = append(errs, missingReferences(l.Snippets[i], folderIDs, tagIDs)...)
	}

	return errors.Join(errs...)
}

func missingReferences(snippet Snippet, folderIDs map[FolderID]bool, tagIDs map[TagID]bool) []error {
	var errs []error

	if !snippet.AtRoot() && !folderIDs[snippet.FolderID()] {
		errs = append(errs, fmt.Errorf("snippet %s: folder %s: %w", snippet.ID(), snippet.FolderID(), ErrMissingFolder))
	}

	for _, tag := range snippet.Tags() {
		if !tagIDs[tag.ID()] {
			errs = append(errs, fmt.Errorf("snippet %s: tag %s: %w", snippet.ID(), tag.ID(), ErrMissingTag))
		}
	}

	return errs
}

func idSet[T, E any](entities []T, id func(T) ID[E]) map[ID[E]]bool {
	ids := make(map[ID[E]]bool, len(entities))
	for _, entity := range entities {
		ids[id(entity)] = true
	}

	return ids
}

func depthOf(folder Folder, parentOf map[FolderID]FolderID) (int, error) {
	visited := map[FolderID]bool{folder.ID(): true}
	depth := 0

	for id := folder.ParentID(); !id.IsNil(); depth++ {
		grandparentID, known := parentOf[id]
		if !known {
			return 0, fmt.Errorf("parent %s: %w", id, ErrMissingFolder)
		}

		if visited[id] {
			return 0, ErrFolderCycle
		}

		visited[id] = true
		id = grandparentID
	}

	return depth, nil
}

func byCreated[T, E any](createdAt func(T) time.Time, id func(T) ID[E]) func(a, b T) int {
	return func(a, b T) int {
		return cmp.Or(createdAt(a).Compare(createdAt(b)), id(a).Compare(id(b)))
	}
}
