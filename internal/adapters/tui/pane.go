package tui

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"

const (
	folderPaneTitle  = "1 Folders"
	tagPaneTitle     = "2 Tags"
	snippetListTitle = "3 " + folderpath.Root + " · by title"
	snippetPaneTitle = "4 Snippet"
)

type pane int

const (
	paneFolders pane = iota
	paneTags
	paneList
	paneSnippet
)

const paneCount = int(paneSnippet) + 1

type navigation func(from, selectionHolder pane) pane

func focusOn(target pane) navigation {
	return func(pane, pane) pane {
		return target
	}
}

func (p pane) next(pane) pane {
	return pane((int(p) + 1) % paneCount)
}

func (p pane) prev(pane) pane {
	return pane((int(p) + paneCount - 1) % paneCount)
}

func (p pane) right(pane) pane {
	if p.inLeftColumn() {
		return paneList
	}

	return paneSnippet
}

func (p pane) left(selectionHolder pane) pane {
	return p.backOut(selectionHolder)
}

func (p pane) drillIn(pane) pane {
	switch p {
	case paneFolders:
		return paneList
	case paneList, paneSnippet:
		return paneSnippet
	case paneTags:
		return paneTags
	}

	return p
}

func (p pane) backOut(selectionHolder pane) pane {
	switch p {
	case paneSnippet:
		return paneList
	case paneList:
		return selectionHolder
	case paneFolders, paneTags:
		return p
	}

	return p
}

func (p pane) inLeftColumn() bool {
	return p == paneFolders || p == paneTags
}
