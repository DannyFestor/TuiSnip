package domain

type Trash struct {
	Folders  map[FolderID]TrashMark
	Snippets map[SnippetID]TrashMark
}
