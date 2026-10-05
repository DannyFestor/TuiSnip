package folderpane

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/nameinput"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Pane struct {
	nameInputKeys binding.Keys
	global        binding.Set
	keys          binding.Set
	rows          []row
	cursor        move.Cursor
	box           look.Size
	naming        naming
	field         nameinput.Field
}

func New(keys binding.Keys) Pane {
	return Pane{
		nameInputKeys: keys,
		global:        keys.For(binding.ScopeGlobal),
		keys:          keys.For(binding.ScopeFolders),
		rows:          rowsOf(browse.Tree{RootSnippetCount: 0, Folders: nil}),
		cursor:        move.Cursor{},
		box:           look.Size{Width: 0, Height: 0},
		naming:        nil,
		field:         nameinput.Field{},
	}
}

func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case look.Resized:
		return p.resized(msg.Box), nil, nil
	case tea.PasteMsg:
		if p.Naming() {
			return p.typed(msg)
		}
	case tea.KeyPressMsg:
		if p.Naming() {
			return p.typed(msg)
		}

		return p.pressed(msg)
	}

	return p, nil, nil
}

func (p Pane) View(frame look.FrameStyle) string {
	lines := p.lines()
	cursor := p.cursor.At(p.highlighted(), len(lines), p.box.Height)

	return cursor.VisibleRows(len(lines), p.box.Height, func(index int) string {
		return look.Row(lines[index].text, lines[index].meta, p.box.Width)
	}, frame.CursorOn)
}

func (p Pane) ShortHelp() []key.Binding {
	if p.Naming() {
		return p.field.ShortHelp()
	}

	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return [][]key.Binding{p.ShortHelp()}
}

func (p Pane) Naming() bool {
	return p.naming != nil
}

func (p Pane) WithTree(tree browse.Tree) Pane {
	next := p
	next.rows = rowsOf(tree)

	return next.withCursor(p.cursor.Index()).WithCursorOn(p.Selected()).withFieldSized()
}

func (p Pane) WithCursorOn(id domain.FolderID) Pane {
	index := slices.IndexFunc(p.rows, func(candidate row) bool { return candidate.folderID == id })
	if index < 0 {
		return p
	}

	return p.withCursor(index)
}

func (p Pane) Selected() domain.FolderID {
	return p.selectedRow().folderID
}

func (p Pane) pressed(msg tea.KeyPressMsg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch {
	case p.keys.Matches(msg, binding.NewFolder):
		return p.startedNaming(newFolder{parentID: p.Selected()}, "")
	case p.keys.Matches(msg, binding.Rename) && !p.Selected().IsNil():
		return p.startedNaming(renamedFolder{folderID: p.Selected()}, p.selectedRow().name)
	case p.keys.Matches(msg, binding.Delete) && !p.Selected().IsNil():
		return p, []outcome.Outcome{outcome.FolderDeleteAsked{ID: p.Selected()}}, nil
	}

	return p.moved(msg)
}

func (p Pane) moved(msg tea.KeyPressMsg) (Pane, []outcome.Outcome, tea.Cmd) {
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

func (p Pane) startedNaming(started naming, initial string) (Pane, []outcome.Outcome, tea.Cmd) {
	next := p
	next.naming = started

	var cmd tea.Cmd

	next.field, cmd = nameinput.New(p.nameInputKeys, validFolderName, initial)

	return next.withFieldSized(), nil, cmd
}

func (p Pane) typed(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	next := p

	var (
		result nameinput.Result
		cmd    tea.Cmd
	)

	next.field, result, cmd = p.field.Update(msg)

	switch result.Ending {
	case nameinput.Committed:
		return next.stoppedNaming(), p.naming.requested(result.Name), cmd
	case nameinput.Cancelled:
		return next.stoppedNaming(), nil, cmd
	case nameinput.Refused:
		return next, []outcome.Outcome{outcome.NoticeShown{Text: refusalText(result.Err)}}, cmd
	case nameinput.Typing:
	}

	return next, nil, cmd
}

func (p Pane) stoppedNaming() Pane {
	p.naming = nil
	p.field = nameinput.Field{}

	return p
}

func (p Pane) resized(box look.Size) Pane {
	next := p
	next.box = box

	return next.withCursor(p.cursor.Index()).withFieldSized()
}

func (p Pane) withFieldSized() Pane {
	if !p.Naming() {
		return p
	}

	p.field = p.field.WithWidth(p.fieldWidth(p.placement()))

	return p
}

func (p Pane) withCursor(index int) Pane {
	p.cursor = p.cursor.At(index, len(p.rows), p.box.Height)

	return p
}

func (p Pane) selectedRow() row {
	return p.rows[p.cursor.Index()]
}

func (p Pane) lines() []line {
	lines := linesOf(p.rows)
	if !p.Naming() {
		return lines
	}

	at := p.placement()

	return at.into(lines, p.field.View())
}

func (p Pane) highlighted() int {
	if !p.Naming() {
		return p.cursor.Index()
	}

	return p.placement().index
}

func (p Pane) placement() placement {
	return p.naming.placed(p.selectedRow(), p.cursor.Index())
}

func (p Pane) fieldWidth(at placement) int {
	return p.box.Width - ansi.StringWidth(at.prefix) - ansi.StringWidth(at.meta) - 1
}
