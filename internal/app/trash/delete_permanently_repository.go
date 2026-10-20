package trash

type DeletePermanentlyRepository interface {
	SnippetDeleter
	FolderDeleter
}
