package tui

const (
	leftColumnPercent    = 22
	snippetListPercent   = 30
	percent              = 100
	focusGrowth          = 8
	focusGrowthPerColumn = focusGrowth / 2
	tallShareNumerator   = 2
	tallShareDenominator = 3
	minimumWidth         = 80
	minimumHeight        = 24
	statusLineHeight     = 1
)

type size struct {
	width  int
	height int
}

type columns struct {
	left    int
	list    int
	snippet int
}

type layout struct {
	single  bool
	folders size
	tags    size
	list    size
	snippet size
}

func arrange(screen size, focus, tallLeft pane) layout {
	panes := size{width: screen.width, height: max(0, screen.height-statusLineHeight)}
	if screen.width < minimumWidth || screen.height < minimumHeight {
		return singlePane(panes)
	}

	widths := columnWidths(panes.width, focus)
	tall := panes.height * tallShareNumerator / tallShareDenominator
	folderHeight, tagHeight := panes.height-tall, tall

	if tallLeft == paneFolders {
		folderHeight, tagHeight = tall, panes.height-tall
	}

	return layout{
		single:  false,
		folders: size{width: widths.left, height: folderHeight},
		tags:    size{width: widths.left, height: tagHeight},
		list:    size{width: widths.list, height: panes.height},
		snippet: size{width: widths.snippet, height: panes.height},
	}
}

func singlePane(panes size) layout {
	return layout{single: true, folders: panes, tags: panes, list: panes, snippet: panes}
}

func columnWidths(width int, focus pane) columns {
	left := width * leftColumnPercent / percent
	list := width * snippetListPercent / percent

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

func (l layout) of(p pane) size {
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

	return size{width: 0, height: 0}
}
