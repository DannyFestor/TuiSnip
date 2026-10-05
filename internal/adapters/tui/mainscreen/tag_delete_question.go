package mainscreen

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

const (
	tagDeleteTitle  = "Delete Tag"
	tagDeleteFormat = "Permanently delete Tag %q and remove it from %s? This cannot be undone."
)

func tagDeleteQuestion(preview tag.DeletePreview) confirm.Question {
	return confirm.Question{
		Title: tagDeleteTitle,
		Text: fmt.Sprintf(
			tagDeleteFormat,
			preview.Tag.Name().String(),
			counted(preview.SnippetCount, "Snippet"),
		),
	}
}
