package trash

import (
	"cmp"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type shelf struct {
	trash           domain.Trash
	folders         map[domain.FolderID]domain.Folder
	subfolders      map[domain.FolderID][]domain.Folder
	snippetsFiled   map[domain.FolderID][]domain.Snippet
	trashedFolders  []domain.Folder
	trashedSnippets []domain.Snippet
}

func shelfOf(loaded domain.Library) shelf {
	stocked := shelf{
		trash:           loaded.Trash,
		folders:         make(map[domain.FolderID]domain.Folder, len(loaded.Folders)),
		subfolders:      make(map[domain.FolderID][]domain.Folder, len(loaded.Folders)),
		snippetsFiled:   make(map[domain.FolderID][]domain.Snippet, len(loaded.Snippets)),
		trashedFolders:  nil,
		trashedSnippets: nil,
	}

	for _, folder := range loaded.Folders {
		stocked = stocked.withFolder(folder)
	}

	for index := range loaded.Snippets {
		stocked = stocked.withSnippet(loaded.Snippets[index])
	}

	return stocked.sorted()
}

func (s shelf) withFolder(folder domain.Folder) shelf {
	s.folders[folder.ID()] = folder

	if _, trashed := s.trash.Folders[folder.ID()]; trashed {
		s.trashedFolders = append(s.trashedFolders, folder)

		return s
	}

	s.subfolders[folder.ParentID()] = append(s.subfolders[folder.ParentID()], folder)

	return s
}

func (s shelf) withSnippet(snippet domain.Snippet) shelf {
	if _, trashed := s.trash.Snippets[snippet.ID()]; trashed {
		s.trashedSnippets = append(s.trashedSnippets, snippet)

		return s
	}

	s.snippetsFiled[snippet.FolderID()] = append(s.snippetsFiled[snippet.FolderID()], snippet)

	return s
}

func (s shelf) sorted() shelf {
	for _, siblings := range s.subfolders {
		slices.SortFunc(siblings, domain.CompareFolders)
	}

	for _, filed := range s.snippetsFiled {
		slices.SortFunc(filed, byTitle)
	}

	return s
}

func (s shelf) items() []Item {
	items := make([]Item, 0, len(s.trashedFolders)+len(s.trashedSnippets))

	for _, folder := range s.trashedFolders {
		items = append(items, s.folderItem(folder))
	}

	for index := range s.trashedSnippets {
		items = append(items, s.snippetItem(s.trashedSnippets[index]))
	}

	slices.SortFunc(items, newestDeletedFirst)

	return items
}

func (s shelf) folderItem(folder domain.Folder) Item {
	mark := s.trash.Folders[folder.ID()]
	outline := s.outlineOf(folder)

	return Item{
		Snippet:   domain.Snippet{},
		Folder:    folder,
		DeletedAt: mark.DeletedAt(),
		Path:      s.pathOf(folder.ParentID(), mark),
		Contents: Contents{
			SubfolderCount: outline.subfolderCount(),
			SnippetCount:   outline.snippetCount(),
			Outline:        outline,
		},
	}
}

func (s shelf) snippetItem(snippet domain.Snippet) Item {
	mark := s.trash.Snippets[snippet.ID()]

	return Item{
		Snippet:   snippet,
		Folder:    domain.Folder{},
		DeletedAt: mark.DeletedAt(),
		Path:      s.pathOf(snippet.FolderID(), mark),
		Contents: Contents{
			SubfolderCount: 0,
			SnippetCount:   0,
			Outline:        Outline{Name: "", SnippetTitles: nil, Subfolders: nil},
		},
	}
}

func (s shelf) outlineOf(folder domain.Folder) Outline {
	filed := s.snippetsFiled[folder.ID()]
	titles := make([]string, 0, len(filed))

	for index := range filed {
		titles = append(titles, filed[index].Title().String())
	}

	subfolders := make([]Outline, 0, len(s.subfolders[folder.ID()]))
	for _, subfolder := range s.subfolders[folder.ID()] {
		subfolders = append(subfolders, s.outlineOf(subfolder))
	}

	return Outline{Name: folder.Name().String(), SnippetTitles: titles, Subfolders: subfolders}
}

func (s shelf) pathOf(folderID domain.FolderID, mark domain.TrashMark) []string {
	if !mark.FormerPath().IsEmpty() {
		return mark.FormerPath().Strings()
	}

	var names []string

	visited := make(map[domain.FolderID]bool)

	for id := folderID; !id.IsNil() && !visited[id]; id = s.folders[id].ParentID() {
		visited[id] = true
		names = append(names, s.folders[id].Name().String())
	}

	slices.Reverse(names)

	return names
}

func (o Outline) subfolderCount() int {
	count := len(o.Subfolders)
	for _, subfolder := range o.Subfolders {
		count += subfolder.subfolderCount()
	}

	return count
}

func (o Outline) snippetCount() int {
	count := len(o.SnippetTitles)
	for _, subfolder := range o.Subfolders {
		count += subfolder.snippetCount()
	}

	return count
}

func byTitle(a, b domain.Snippet) int {
	return cmp.Or(a.Title().CompareIgnoringCase(b.Title()), a.ID().Compare(b.ID()))
}

func newestDeletedFirst(a, b Item) int {
	return cmp.Or(b.DeletedAt.Compare(a.DeletedAt), b.ID().compare(a.ID()))
}
