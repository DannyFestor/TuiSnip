package mainscreen

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	snippetDeleteTitle  = "Delete Snippet"
	snippetDeleteFormat = "Permanently delete Snippet %q? This cannot be undone."
)

func snippetDeleteQuestion(snippet domain.Snippet) confirm.Question {
	return confirm.Question{
		Title: snippetDeleteTitle,
		Text:  fmt.Sprintf(snippetDeleteFormat, snippet.Title().String()),
	}
}
