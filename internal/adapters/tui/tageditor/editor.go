package tageditor

import (
	"cmp"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagrefusal"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	editorPercent = 60
	untitled      = ""
	filterPrompt  = "filter or new Tag: "
	noTags        = "No Tags yet."
	headerLabel   = "Tags  "
	nameSeparator = ", "
	headerRows    = 1
)

type Editor struct {
	hints  []key.Binding
	styles look.Styles
	listed []browse.TagCount
	state  toggles
	rows   []row
	list   picker.List
	outer  look.Size
}

func New(keys binding.Keys, styles look.Styles, offer Offer) (Editor, tea.Cmd) {
	list, cmd := picker.New(keys, styles, picker.Labels{Prompt: filterPrompt, NoMatches: noTags}, nil)
	opened := Editor{
		hints:  keys.TagEditorHints(),
		styles: styles,
		listed: offer.Listed,
		state:  togglesFrom(offer.Chosen),
		rows:   nil,
		list:   list,
		outer:  look.Size{Width: 0, Height: 0},
	}

	return opened.relisted(), cmd
}

func (e Editor) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case look.Resized:
		return outcome.Stay(e.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(e.restyled(msg))
	case arrived.Tags:
		return outcome.Stay(e.withListed(msg.Tags))
	}

	return e.listUpdated(msg)
}

func (e Editor) View() string {
	return look.Frame(e.styles.Focused, untitled, e.header()+"\n"+e.list.View(), e.outer)
}

func (e Editor) ShortHelp() []key.Binding {
	return slices.Clone(e.hints)
}

func (e Editor) FullHelp() [][]key.Binding {
	return [][]key.Binding{e.ShortHelp()}
}

func (e Editor) listUpdated(msg tea.Msg) outcome.Step {
	next := e

	var (
		result picker.Result
		cmd    tea.Cmd
	)

	next.list, result, cmd = e.list.Update(msg)

	switch result.Ending {
	case picker.Picked:
		return next.toggled(e.rows[result.Index]).Running(cmd)
	case picker.Cancelled:
		return outcome.Close().Passing(outcome.TagsEdited{Chosen: e.state.chosen}).Running(cmd)
	case picker.Filtering:
	}

	if next.list.Filter() != e.list.Filter() {
		next = next.relisted()
	}

	return outcome.Stay(next).Running(cmd)
}

func (e Editor) toggled(picked row) outcome.Step {
	state, err := picked.toggledIn(e.state)
	if err != nil {
		return outcome.Stay(e).Passing(outcome.NoticeShown{Text: tagrefusal.Text(err)})
	}

	next := e
	next.state = state
	next = next.relisted()
	next.list = next.list.WithCursorOn(picked.landing())

	return outcome.Stay(next)
}

func (e Editor) withListed(listed []browse.TagCount) Editor {
	next := e
	next.listed = listed

	return next.relisted()
}

func (e Editor) relisted() Editor {
	e.rows = e.withCreateRow(e.tagRows())
	e.list = e.list.WithChoices(e.choices())

	return e
}

func (e Editor) tagRows() []row {
	rows := make([]row, 0, len(e.listed)+len(e.state.offered))
	for _, counted := range e.listed {
		rows = append(rows, storedRow{counted: counted})
	}

	for _, name := range e.state.offered {
		rows = append(rows, newRow{name: name})
	}

	slices.SortStableFunc(rows, func(a, b row) int { return cmp.Compare(a.key(), b.key()) })

	return rows
}

func (e Editor) withCreateRow(rows []row) []row {
	typedKey := value.TagNameKey(e.list.Filter())
	named := func(listed row) bool { return listed.key() == typedKey }

	if typedKey == "" || slices.ContainsFunc(rows, named) {
		return rows
	}

	return append(rows, createRow{typed: e.list.Filter()})
}

func (e Editor) choices() []picker.Choice {
	choices := make([]picker.Choice, 0, len(e.rows))
	for _, listed := range e.rows {
		choices = append(choices, listed.choice(e.state))
	}

	return choices
}

func (e Editor) header() string {
	return headerLabel + strings.Join(e.state.chosen.Names(), nameSeparator)
}

func (e Editor) resized(screen look.Size) Editor {
	next := e
	next.outer = screen.Share(editorPercent)

	inner := next.outer.Inner()
	next.list, _, _ = e.list.Update(look.Resized{Box: look.Size{Width: inner.Width, Height: inner.Height - headerRows}})

	return next
}

func (e Editor) restyled(msg look.Restyled) Editor {
	e.styles = msg.Styles
	e.list, _, _ = e.list.Update(msg)

	return e
}
