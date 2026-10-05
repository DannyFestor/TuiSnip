package snippetlist

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type List struct {
	global     binding.Set
	keys       binding.Set
	emptyHints []key.Binding
	styles     look.Styles
	meta       Meta
	snippets   []domain.Snippet
	cursor     move.Cursor
	box        look.Size
}

func New(keys binding.Keys, styles look.Styles, meta Meta) List {
	return List{
		global:     keys.For(binding.ScopeGlobal),
		keys:       keys.For(binding.ScopeSnippetList),
		emptyHints: keys.EmptyListHints(),
		styles:     styles,
		meta:       meta,
		snippets:   nil,
		cursor:     move.Cursor{},
		box:        look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (l List) Update(msg tea.Msg) (List, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return l.pressed(msg), l.copyRequested(msg), nil
	case look.Resized:
		return l.resized(msg.Box), nil, nil
	}

	return l, nil, nil
}

func (l List) View(frame look.FrameStyle) string {
	if len(l.snippets) == 0 {
		return look.EmptyHint(l.styles, l.emptyHints)
	}

	return l.rows(frame)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithSnippets(snippets []domain.Snippet) List {
	next := l
	next.snippets = snippets

	return next.withCursor(l.cursor.Index())
}

func (l List) WithMeta(meta Meta) List {
	l.meta = meta

	return l
}

func (l List) WithCursorOn(id domain.SnippetID) List {
	index := slices.IndexFunc(l.snippets, func(candidate domain.Snippet) bool { return candidate.ID() == id })
	if index < 0 {
		return l
	}

	return l.withCursor(index)
}

func (l List) Moved(direction move.Direction) List {
	l.cursor = l.cursor.Moved(direction, len(l.snippets), l.box.Height)

	return l
}

func (l List) Selected() (domain.Snippet, bool) {
	if len(l.snippets) == 0 {
		return domain.Snippet{}, false
	}

	return l.snippets[l.cursor.Index()], true
}

func (l List) Snippets() []domain.Snippet {
	return l.snippets
}

func (l List) pressed(msg tea.KeyPressMsg) List {
	direction, ok := move.Pressed(l.global, msg)
	if !ok {
		return l
	}

	return l.Moved(direction)
}

func (l List) copyRequested(msg tea.KeyPressMsg) []outcome.Outcome {
	selected, ok := l.Selected()
	if !ok || !l.keys.Matches(msg, binding.Copy) {
		return nil
	}

	return []outcome.Outcome{outcome.CopyRequested{ID: selected.ID()}}
}

func (l List) resized(box look.Size) List {
	next := l
	next.box = box

	return next.withCursor(l.cursor.Index())
}

func (l List) rows(frame look.FrameStyle) string {
	return l.cursor.VisibleRows(len(l.snippets), l.box.Height, l.rowAt, frame.CursorOn)
}

func (l List) rowAt(index int) string {
	snippet := l.snippets[index]

	return look.Row(snippet.Title().String(), l.meta(snippet), l.box.Width)
}

func (l List) withCursor(index int) List {
	l.cursor = l.cursor.At(index, len(l.snippets), l.box.Height)

	return l
}
