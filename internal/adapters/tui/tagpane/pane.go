package tagpane

import (
	"slices"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	noTagsText = "No Tags yet."
	TagPrefix  = "# "
)

type Pane struct {
	global binding.Set
	keys   binding.Set
	styles look.Styles
	tags   []browse.TagCount
	cursor move.Cursor
	box    look.Size
}

func New(keys binding.Keys, styles look.Styles) Pane {
	return Pane{
		global: keys.For(binding.ScopeGlobal),
		keys:   keys.For(binding.ScopeTags),
		styles: styles,
		tags:   nil,
		cursor: move.Cursor{},
		box:    look.Size{Width: 0, Height: 0},
	}
}

func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case look.Resized:
		return p.resized(msg.Box), nil, nil
	case look.Restyled:
		return p.restyled(msg.Styles), nil, nil
	case tea.KeyPressMsg:
		return p.moved(msg)
	}

	return p, nil, nil
}

func (p Pane) View(frame look.FrameStyle) string {
	if len(p.tags) == 0 {
		return look.FitWidth(p.styles.Dim.Render(noTagsText), p.box.Width)
	}

	return p.cursor.VisibleRows(len(p.tags), p.box.Height, func(index int) string {
		counted := p.tags[index]

		return look.Row(TagPrefix+counted.Tag.Name().String(), strconv.Itoa(counted.SnippetCount), p.box.Width)
	}, frame.CursorOn)
}

func (p Pane) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}

func (p Pane) WithTags(tags []browse.TagCount) Pane {
	selected, _ := p.Selected()

	next := p
	next.tags = tags

	index := slices.IndexFunc(tags, func(counted browse.TagCount) bool { return counted.Tag.ID() == selected })
	if index < 0 {
		index = p.cursor.Index()
	}

	return next.withCursor(index)
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

func (p Pane) resized(box look.Size) Pane {
	next := p
	next.box = box

	return next.withCursor(p.cursor.Index())
}

func (p Pane) restyled(styles look.Styles) Pane {
	p.styles = styles

	return p
}

func (p Pane) withCursor(index int) Pane {
	p.cursor = p.cursor.At(index, len(p.tags), p.box.Height)

	return p
}
