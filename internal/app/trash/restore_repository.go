package trash

type RestoreRepository interface {
	SnippetRestorer
	FolderRestorer
}
