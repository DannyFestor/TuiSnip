package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const snippetListTitle = "3 " + rootPath + " · by title"

type snippetMeta func(domain.Snippet) string

type snippetList struct {
	snippets []domain.Snippet
	meta     snippetMeta
	cursor   int
	offset   int
	height   int
}

func newSnippetList(meta snippetMeta) snippetList {
	return snippetList{snippets: nil, meta: meta, cursor: 0, offset: 0, height: 0}
}

func (l snippetList) withSnippets(snippets []domain.Snippet) snippetList {
	next := l
	next.snippets = snippets

	return next.withCursor(l.cursor)
}

func (l snippetList) withCursorOn(id domain.SnippetID) snippetList {
	index := slices.IndexFunc(l.snippets, func(candidate domain.Snippet) bool { return candidate.ID() == id })
	if index < 0 {
		return l
	}

	return l.withCursor(index)
}

func (l snippetList) resized(height int) snippetList {
	next := l
	next.height = height

	return next.withCursor(l.cursor)
}

func (l snippetList) moved(direction move.Direction) snippetList {
	return l.withCursor(l.cursorAfter(direction))
}

func (l snippetList) selected() (domain.Snippet, bool) {
	if len(l.snippets) == 0 {
		return domain.Snippet{}, false
	}

	return l.snippets[l.cursor], true
}

func (l snippetList) body(styles look.Styles, paneStyle look.FrameStyle, width int, hints []key.Binding) string {
	if len(l.snippets) == 0 {
		return emptyHint(styles, hints)
	}

	return l.rows(paneStyle, width)
}

func (l snippetList) rows(paneStyle look.FrameStyle, width int) string {
	end := min(len(l.snippets), l.offset+l.height)
	rows := make([]string, 0, max(0, end-l.offset))

	for index := l.offset; index < end; index++ {
		rows = append(rows, l.rowAt(paneStyle, index, width))
	}

	return strings.Join(rows, "\n")
}

func (l snippetList) rowAt(paneStyle look.FrameStyle, index, width int) string {
	snippet := l.snippets[index]
	line := look.Row(snippet.Title().String(), l.meta(snippet), width)

	if index == l.cursor {
		return paneStyle.Cursor.Render(line)
	}

	return line
}

func (l snippetList) cursorAfter(direction move.Direction) int {
	switch direction {
	case move.Down:
		return l.cursor + 1
	case move.Up:
		return l.cursor - 1
	case move.Top:
		return 0
	case move.Bottom:
		return len(l.snippets) - 1
	case move.PageDown:
		return l.cursor + l.height
	case move.PageUp:
		return l.cursor - l.height
	case move.None:
		return l.cursor
	}

	return l.cursor
}

func (l snippetList) withCursor(cursor int) snippetList {
	l.cursor = max(0, min(cursor, len(l.snippets)-1))
	l.offset = max(0, min(l.offset, l.cursor), l.cursor-l.height+1)

	return l
}

func languageOf(snippet domain.Snippet) string {
	return snippet.FirstFragment().Language().String()
}

func folderPathOf(domain.Snippet) string {
	return rootPath
}
