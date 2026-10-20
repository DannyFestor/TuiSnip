package trashlist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	folderMarker = "▸ "
	snippetNoun  = "Snippet"
)

type List struct {
	global          binding.Set
	keys            binding.Set
	styles          look.Styles
	snippetLanguage func(domain.Snippet) string
	items           []trash.Item
	cursor          move.Cursor
	box             look.Size
}

func New(keys binding.Keys, styles look.Styles, snippetLanguage func(domain.Snippet) string) List {
	return List{
		global:          keys.For(binding.ScopeGlobal),
		keys:            keys.For(binding.ScopeTrash),
		styles:          styles,
		snippetLanguage: snippetLanguage,
		items:           nil,
		cursor:          move.Cursor{},
		box:             look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (l List) Update(msg tea.Msg) (List, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return l.pressed(msg), nil, nil
	case look.Resized:
		return l.resized(msg.Box), nil, nil
	case look.Restyled:
		l.styles = msg.Styles

		return l, nil, nil
	case pointer.Clicked:
		l.cursor, _ = l.cursor.ClickedAt(msg.At.Y, len(l.items), l.box.Height)

		return l, nil, nil
	case pointer.Wheeled:
		l.cursor = l.cursor.Scrolled(msg.Lines, len(l.items), l.box.Height)

		return l, nil, nil
	}

	return l, nil, nil
}

func (l List) View(frame look.FrameStyle) string {
	if len(l.items) == 0 {
		return look.EmptyTrash(l.styles)
	}

	return l.cursor.VisibleRows(len(l.items), l.box.Height, l.rowAt, frame.CursorOn)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithItems(items []trash.Item) List {
	next := l
	next.items = items

	return next.withCursor(l.cursor.Index())
}

func (l List) Selected() (trash.Item, bool) {
	if len(l.items) == 0 {
		return trash.Item{}, false
	}

	return l.items[l.cursor.Index()], true
}

func (l List) HasRowAt(at pointer.Point) bool {
	_, ok := l.cursor.RowAt(at.Y, len(l.items), l.box.Height)

	return ok
}

func (l List) pressed(msg tea.KeyPressMsg) List {
	direction, ok := move.Pressed(l.global, msg)
	if !ok {
		return l
	}

	l.cursor = l.cursor.Moved(direction, len(l.items), l.box.Height)

	return l
}

func (l List) resized(box look.Size) List {
	next := l
	next.box = box

	return next.withCursor(l.cursor.Index())
}

func (l List) withCursor(index int) List {
	l.cursor = l.cursor.At(index, len(l.items), l.box.Height)

	return l
}

func (l List) rowAt(index int) string {
	item := l.items[index]
	if item.IsFolder() {
		return look.Row(folderMarker+item.Name(), look.Counted(item.Contents.SnippetCount, snippetNoun), l.box.Width)
	}

	return look.Row(item.Name(), l.snippetLanguage(item.Snippet), l.box.Width)
}
