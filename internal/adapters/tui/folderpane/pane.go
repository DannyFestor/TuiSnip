package folderpane

import (
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Pane struct {
	global binding.Set
	keys   binding.Set
	rows   []row
	cursor move.Cursor
	box    look.Size
}

func New(keys binding.Keys) Pane {
	return Pane{
		global: keys.For(binding.ScopeGlobal),
		keys:   keys.For(binding.ScopeFolders),
		rows:   rowsOf(browse.Tree{RootSnippetCount: 0, Folders: nil}),
		cursor: move.Cursor{},
		box:    look.Size{Width: 0, Height: 0},
	}
}

func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg)
	case look.Resized:
		return p.resized(msg.Box), nil, nil
	}

	return p, nil, nil
}

func (p Pane) View(frame look.FrameStyle) string {
	end := p.cursor.End(len(p.rows), p.box.Height)
	lines := make([]string, 0, end-p.cursor.Offset())

	for index := p.cursor.Offset(); index < end; index++ {
		lines = append(lines, p.rowAt(frame, index))
	}

	return strings.Join(lines, "\n")
}

func (p Pane) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}

func (p Pane) WithTree(tree browse.Tree) Pane {
	next := p
	next.rows = rowsOf(tree)

	return next.withCursor(p.cursor.Index()).WithCursorOn(p.Selected())
}

func (p Pane) WithCursorOn(id domain.FolderID) Pane {
	index := slices.IndexFunc(p.rows, func(candidate row) bool { return candidate.folderID == id })
	if index < 0 {
		return p
	}

	return p.withCursor(index)
}

func (p Pane) Selected() domain.FolderID {
	return p.rows[p.cursor.Index()].folderID
}

func (p Pane) pressed(msg tea.KeyPressMsg) (Pane, []outcome.Outcome, tea.Cmd) {
	direction, ok := move.Pressed(p.global, msg)
	if !ok {
		return p, nil, nil
	}

	next := p
	next.cursor = p.cursor.Moved(direction, len(p.rows), p.box.Height)

	if next.Selected() == p.Selected() {
		return next, nil, nil
	}

	return next, []outcome.Outcome{outcome.FolderSelected{ID: next.Selected()}}, nil
}

func (p Pane) resized(box look.Size) Pane {
	next := p
	next.box = box

	return next.withCursor(p.cursor.Index())
}

func (p Pane) withCursor(index int) Pane {
	p.cursor = p.cursor.At(index, len(p.rows), p.box.Height)

	return p
}

func (p Pane) rowAt(frame look.FrameStyle, index int) string {
	shown := p.rows[index]
	line := look.Row(shown.label, strconv.Itoa(shown.snippetCount), p.box.Width)

	if index == p.cursor.Index() {
		return frame.Cursor.Render(line)
	}

	return line
}
