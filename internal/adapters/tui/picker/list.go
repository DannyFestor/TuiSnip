package picker

import (
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
)

const filterRows = 2

type List struct {
	keys      binding.Set
	labels    Labels
	filter    textinput.Model
	choices   []Choice
	greyedOut []int
	shown     []int
	cursor    move.Cursor
	box       look.Size
}

func New(keys binding.Keys, labels Labels, choices []Choice) (List, tea.Cmd) {
	filter := input.NewLine(labels.Prompt)
	cmd := filter.Focus()

	list := List{
		keys:      keys.For(binding.ScopePicker),
		labels:    labels,
		filter:    filter,
		choices:   choices,
		greyedOut: nil,
		shown:     nil,
		cursor:    move.Cursor{},
		box:       look.Size{Width: 0, Height: 0},
	}

	return list.filtered(), cmd
}

func (l List) Update(msg tea.Msg) (List, Result, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return l.pressed(msg)
	case look.Resized:
		return l.resized(msg.Box), filtering(), nil
	}

	return l.typed(msg)
}

func (l List) View(styles look.Styles) string {
	return l.filter.View() + "\n\n" + l.rows(styles)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithChoices(choices []Choice) List {
	next := l
	next.choices = choices

	return next.filtered()
}

func (l List) WithGreyedOut(indexes ...int) List {
	l.greyedOut = indexes

	return l
}

func (l List) WithCursorOn(text string) List {
	return l.cursorOnShown(slices.IndexFunc(l.shown, func(index int) bool { return l.choices[index].Text == text }))
}

func (l List) WithCursorOnChoice(index int) List {
	return l.cursorOnShown(slices.Index(l.shown, index))
}

func (l List) Highlighted() (Choice, bool) {
	if len(l.shown) == 0 {
		return Choice{Text: "", Meta: ""}, false
	}

	return l.choices[l.shown[l.cursor.Index()]], true
}

func (l List) pressed(msg tea.KeyPressMsg) (List, Result, tea.Cmd) {
	switch {
	case l.keys.Matches(msg, binding.Down):
		return l.moved(move.Down), filtering(), nil
	case l.keys.Matches(msg, binding.Up):
		return l.moved(move.Up), filtering(), nil
	case l.keys.Matches(msg, binding.Accept):
		return l, l.picked(), nil
	case l.keys.Matches(msg, binding.Cancel):
		return l, cancelled(), nil
	}

	return l.typed(msg)
}

func (l List) picked() Result {
	if len(l.shown) == 0 || l.greyed(l.cursor.Index()) {
		return filtering()
	}

	return pickedAt(l.shown[l.cursor.Index()])
}

func (l List) greyed(shownAt int) bool {
	return slices.Contains(l.greyedOut, l.shown[shownAt])
}

func (l List) typed(msg tea.Msg) (List, Result, tea.Cmd) {
	next := l

	var cmd tea.Cmd

	next.filter, cmd = l.filter.Update(msg)
	if next.filter.Value() == l.filter.Value() {
		return next, filtering(), cmd
	}

	return next.filtered(), filtering(), cmd
}

func (l List) filtered() List {
	next := l
	next.shown = matching(l.choices, l.filter.Value())

	return next.withCursor(0)
}

func (l List) moved(direction move.Direction) List {
	l.cursor = l.cursor.Moved(direction, len(l.shown), l.rowsHeight())

	return l
}

func (l List) cursorOnShown(at int) List {
	if at < 0 {
		return l
	}

	return l.withCursor(at)
}

func (l List) withCursor(index int) List {
	l.cursor = l.cursor.At(index, len(l.shown), l.rowsHeight())

	return l
}

func (l List) resized(box look.Size) List {
	next := l
	next.box = box
	next.filter.SetWidth(max(1, box.Width-ansi.StringWidth(l.labels.Prompt)-input.CursorWidth))

	return next.withCursor(l.cursor.Index())
}

func (l List) rowsHeight() int {
	return max(0, l.box.Height-filterRows)
}

func (l List) rows(styles look.Styles) string {
	if len(l.shown) == 0 {
		return styles.Dim.Render(l.labels.NoMatches)
	}

	return l.cursor.VisibleRows(len(l.shown), l.rowsHeight(), func(at int) string {
		choice := l.choices[l.shown[at]]

		return l.rowStyle(at, styles).Render(look.Row(choice.Text, choice.Meta, l.box.Width))
	}, func(line string) string { return l.cursorStyle(styles).Render(line) })
}

// The cursor style wraps the cursor's row, so that row stays unstyled until then.
func (l List) rowStyle(at int, styles look.Styles) lipgloss.Style {
	if at != l.cursor.Index() && l.greyed(at) {
		return styles.Dim
	}

	return styles.Plain
}

func (l List) cursorStyle(styles look.Styles) lipgloss.Style {
	if l.greyed(l.cursor.Index()) {
		return styles.Focused.Cursor.Foreground(styles.Dim.GetForeground())
	}

	return styles.Focused.Cursor
}
