package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

const (
	leftColumnPercent    = 22
	snippetListPercent   = 30
	focusGrowth          = 8
	focusGrowthPerColumn = focusGrowth / 2
	tallShareNumerator   = 2
	tallShareDenominator = 3
	minimumWidth         = 80
	minimumHeight        = 24
	statusLineHeight     = 1
)

type columns struct {
	left    int
	list    int
	snippet int
}

type layout struct {
	single  bool
	folders look.Size
	tags    look.Size
	list    look.Size
	snippet look.Size
}

func arrange(screen look.Size, focus, tallLeft pane) layout {
	if tooSmall(screen) {
		return singlePane(screen)
	}

	panes := paneArea(screen)
	widths := columnWidths(panes.Width, focus)
	tall := panes.Height * tallShareNumerator / tallShareDenominator
	folderHeight, tagHeight := panes.Height-tall, tall

	if tallLeft == paneFolders {
		folderHeight, tagHeight = tall, panes.Height-tall
	}

	return layout{
		single:  false,
		folders: look.Size{Width: widths.left, Height: folderHeight},
		tags:    look.Size{Width: widths.left, Height: tagHeight},
		list:    look.Size{Width: widths.list, Height: panes.Height},
		snippet: look.Size{Width: widths.snippet, Height: panes.Height},
	}
}

func tooSmall(screen look.Size) bool {
	return screen.Width < minimumWidth || screen.Height < minimumHeight
}

func singlePane(screen look.Size) layout {
	panes := paneArea(screen)

	return layout{single: true, folders: panes, tags: panes, list: panes, snippet: panes}
}

func paneArea(screen look.Size) look.Size {
	return look.Size{Width: screen.Width, Height: max(0, screen.Height-statusLineHeight)}
}

func columnWidths(width int, focus pane) columns {
	left := width * leftColumnPercent / look.Percent
	list := width * snippetListPercent / look.Percent

	switch focus {
	case paneFolders, paneTags:
		left += focusGrowth
	case paneList:
		list += focusGrowth
	case paneSnippet:
		left -= focusGrowthPerColumn
		list -= focusGrowthPerColumn
	}

	return columns{left: left, list: list, snippet: width - left - list}
}

func (l layout) of(p pane) look.Size {
	switch p {
	case paneFolders:
		return l.folders
	case paneTags:
		return l.tags
	case paneList:
		return l.list
	case paneSnippet:
		return l.snippet
	}

	return look.Size{Width: 0, Height: 0}
}
