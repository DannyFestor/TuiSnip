package mainscreen

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetsLoaded struct {
	Snippets  []domain.Snippet
	Selecting domain.SnippetID
}
