package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const snippetListTitle = "3 " + rootPath + " · by title"

type snippetList struct {
	snippets []domain.Snippet
	cursor   int
	offset   int
	height   int
}

func (l snippetList) withSnippets(snippets []domain.Snippet) snippetList {
	next := l
	next.snippets = snippets

	return next.withCursor(l.cursor)
}

func (l snippetList) resized(height int) snippetList {
	next := l
	next.height = height

	return next.withCursor(l.cursor)
}

func (l snippetList) moved(move movement) snippetList {
	return l.withCursor(l.cursorAfter(move))
}

func (l snippetList) selected() (domain.Snippet, bool) {
	if len(l.snippets) == 0 {
		return domain.Snippet{}, false
	}

	return l.snippets[l.cursor], true
}

func (l snippetList) body(styles styleSet, look paneLook, width int, hints []key.Binding) string {
	if len(l.snippets) == 0 {
		return emptyHint(styles, hints)
	}

	end := min(len(l.snippets), l.offset+l.height)
	rows := make([]string, 0, end-l.offset)

	for index := l.offset; index < end; index++ {
		rows = append(rows, l.rowAt(look, index, width))
	}

	return strings.Join(rows, "\n")
}

func (l snippetList) rowAt(look paneLook, index, width int) string {
	snippet := l.snippets[index]
	line := row(snippet.Title().String(), snippet.FirstFragment().Language().String(), width)

	if index == l.cursor {
		return look.cursor.Render(line)
	}

	return line
}

func (l snippetList) cursorAfter(move movement) int {
	switch move {
	case moveDown:
		return l.cursor + 1
	case moveUp:
		return l.cursor - 1
	case moveTop:
		return 0
	case moveBottom:
		return len(l.snippets) - 1
	case movePageDown:
		return l.cursor + l.height
	case movePageUp:
		return l.cursor - l.height
	case moveNone:
		return l.cursor
	}

	return l.cursor
}

func (l snippetList) withCursor(cursor int) snippetList {
	l.cursor = max(0, min(cursor, len(l.snippets)-1))
	l.offset = max(0, min(l.offset, l.cursor), l.cursor-l.height+1)

	return l
}
