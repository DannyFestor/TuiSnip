package tui

import "strconv"

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

func (f folderPane) body(look paneLook, width int) string {
	return look.cursor.Render(row(rootLabel, strconv.Itoa(f.rootSnippetCount), width))
}
