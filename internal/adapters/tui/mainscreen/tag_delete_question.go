package mainscreen

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

const (
	tagDeleteTitle             = "Delete Tag"
	tagDeleteFormatPlaceholder = "PLACEHOLDER, wording pending on #111: delete Tag %q from %s?"
)

func tagDeleteQuestion(preview tag.DeletePreview) confirm.Question {
	return confirm.Question{
		Title: tagDeleteTitle,
		Text: fmt.Sprintf(
			tagDeleteFormatPlaceholder,
			preview.Tag.Name().String(),
			counted(preview.SnippetCount, "Snippet"),
		),
	}
}
