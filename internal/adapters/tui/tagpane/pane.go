package tagpane

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

const (
	noTagsText = "No Tags yet."
	TagPrefix  = "# "
)

type Pane struct {
	nameInputKeys binding.Keys
	global        binding.Set
	keys          binding.Set
	styles        look.Styles
	tags          []browse.TagCount
	cursor        move.Cursor
	box           look.Size
	naming        naming
	field         nameinput.Field
}

func New(keys binding.Keys, styles look.Styles) Pane {
	return Pane{
		nameInputKeys: keys,
		global:        keys.For(binding.ScopeGlobal),
		keys:          keys.For(binding.ScopeTags),
		styles:        styles,
		tags:          nil,
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
	case look.Restyled:
		return p.restyled(msg.Styles), nil, nil
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
	if len(lines) == 0 {
		return look.FitWidth(p.styles.Dim.Render(noTagsText), p.box.Width)
	}

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
	return p.keys.FullHelp()
}

func (p Pane) Naming() bool {
	return p.naming != nil
}

func (p Pane) Tags() []browse.TagCount {
	return slices.Clone(p.tags)
}

func (p Pane) WithTags(tags []browse.TagCount) Pane {
	selected, _ := p.Selected()

	next := p
	next.tags = tags

	index := slices.IndexFunc(tags, func(counted browse.TagCount) bool { return counted.Tag.ID() == selected })
	if index < 0 {
		index = p.cursor.Index()
	}

	return next.withCursor(index).withFieldSized()
}

func (p Pane) WithCursorOn(id domain.TagID) Pane {
	index := slices.IndexFunc(p.tags, func(counted browse.TagCount) bool { return counted.Tag.ID() == id })
	if index < 0 {
		return p
	}

	return p.withCursor(index)
}

func (p Pane) Selected() (domain.TagID, bool) {
	if len(p.tags) == 0 {
		return domain.TagID{}, false
	}

	return p.tags[p.cursor.Index()].Tag.ID(), true
}

func (p Pane) SelectedName() string {
	if len(p.tags) == 0 {
		return ""
	}

	return p.tags[p.cursor.Index()].Tag.Name().String()
}

func (p Pane) SelectedTag() (domain.Tag, bool) {
	if len(p.tags) == 0 {
		return domain.Tag{}, false
	}

	return p.tags[p.cursor.Index()].Tag, true
}

func (p Pane) pressed(msg tea.KeyPressMsg) (Pane, []outcome.Outcome, tea.Cmd) {
	selected, onTag := p.Selected()

	switch {
	case p.keys.Matches(msg, binding.NewTag):
		return p.startedNaming(newTag{existing: p.tags}, "")
	case onTag && p.keys.Matches(msg, binding.Rename):
		under := p.tags[p.cursor.Index()]

		return p.startedNaming(renamedTag{tagID: selected, snippetCount: under.SnippetCount}, p.SelectedName())
	case onTag && p.keys.Matches(msg, binding.Delete):
		return p, []outcome.Outcome{outcome.TagDeleteAsked{ID: selected}}, nil
	}

	return p.moved(msg)
}

func (p Pane) moved(msg tea.KeyPressMsg) (Pane, []outcome.Outcome, tea.Cmd) {
	direction, pressed := move.Pressed(p.global, msg)
	if !pressed {
		return p, nil, nil
	}

	before, _ := p.Selected()

	next := p
	next.cursor = p.cursor.Moved(direction, len(p.tags), p.box.Height)

	after, onTag := next.Selected()
	if !onTag || after == before {
		return next, nil, nil
	}

	return next, []outcome.Outcome{outcome.TagSelected{ID: after}}, nil
}

func (p Pane) startedNaming(started naming, initial string) (Pane, []outcome.Outcome, tea.Cmd) {
	next := p
	next.naming = started

	var cmd tea.Cmd

	next.field, cmd = nameinput.New(p.nameInputKeys, started.validated, initial)

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

func (p Pane) restyled(styles look.Styles) Pane {
	p.styles = styles

	return p
}

func (p Pane) withCursor(index int) Pane {
	p.cursor = p.cursor.At(index, len(p.tags), p.box.Height)

	return p
}

func (p Pane) withFieldSized() Pane {
	if !p.Naming() {
		return p
	}

	p.field = p.field.WithWidth(p.fieldWidth())

	return p
}

func (p Pane) lines() []line {
	lines := linesOf(p.tags)
	if !p.Naming() {
		return lines
	}

	return p.placement().into(lines, p.field.View())
}

func (p Pane) highlighted() int {
	if !p.Naming() {
		return p.cursor.Index()
	}

	return p.placement().index
}

func (p Pane) placement() placement {
	return p.naming.placed(p.cursor.Index(), p.tags, p.field.Typed())
}

func (p Pane) fieldWidth() int {
	return p.box.Width - ansi.StringWidth(TagPrefix) - ansi.StringWidth(p.placement().meta) - 1
}
