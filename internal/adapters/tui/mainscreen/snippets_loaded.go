package mainscreen

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetsLoaded struct {
	FolderID  domain.FolderID
	Snippets  []domain.Snippet
	Selecting domain.SnippetID
}
