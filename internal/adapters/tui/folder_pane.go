package tui

import (
	"strconv"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

const (
	folderPaneTitle = "1 Folders"
	rootLabel       = "◆ Root"
	rootPath        = "Root"
)

type folderPane struct {
	rootSnippetCount int
}

func (f folderPane) withRootSnippetCount(count int) folderPane {
	f.rootSnippetCount = count

	return f
}

func (f folderPane) body(paneStyle look.FrameStyle, width int) string {
	return paneStyle.Cursor.Render(look.Row(rootLabel, strconv.Itoa(f.rootSnippetCount), width))
}
