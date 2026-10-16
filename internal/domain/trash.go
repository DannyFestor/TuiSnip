package domain

type Trash struct {
	Folders  map[FolderID]TrashMark
	Snippets map[SnippetID]TrashMark
}

func (t Trash) holdsFolderOrAncestor(folderID FolderID, parentOf map[FolderID]FolderID) bool {
	visited := make(map[FolderID]bool)

	for id := folderID; !id.IsNil() && !visited[id]; id = parentOf[id] {
		if _, trashed := t.Folders[id]; trashed {
			return true
		}

		visited[id] = true
	}

	return false
}
