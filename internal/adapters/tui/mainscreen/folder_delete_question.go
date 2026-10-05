package mainscreen

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
)

const (
	folderDeleteTitle  = "Delete Folder"
	folderDeleteFormat = "Permanently delete Folder %q, %s, and %s? This cannot be undone."
)

func folderDeleteQuestion(preview folder.DeletePreview) confirm.Question {
	return confirm.Question{
		Title: folderDeleteTitle,
		Text: fmt.Sprintf(
			folderDeleteFormat,
			preview.Folder.Name().String(),
			counted(preview.SubfolderCount, "subfolder"),
			counted(preview.SnippetCount, "Snippet"),
		),
	}
}

func counted(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}

	return fmt.Sprintf("%d %ss", count, noun)
}
