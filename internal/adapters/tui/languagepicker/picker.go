package languagepicker

import (
	"cmp"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	pickerPercent = 60
	filterPrompt  = "filter: "
	noMatches     = "No Language matches."
)

type Picker struct {
	keys       binding.Set
	styles     look.Styles
	offer      Offer
	showingAll bool
	listed     []value.Language
	list       picker.List
	outer      look.Size
}

func New(keys binding.Keys, styles look.Styles, offer Offer) (Picker, tea.Cmd) {
	opened := Picker{
		keys:       keys.For(binding.ScopePicker),
		styles:     styles,
		offer:      offer,
		showingAll: len(offer.Curated) == 0,
		listed:     nil,
		list:       picker.List{},
		outer:      look.Size{Width: 0, Height: 0},
	}
	opened.listed = opened.offered()

	var cmd tea.Cmd

	opened.list, cmd = picker.New(
		keys,
		picker.Labels{Prompt: filterPrompt, NoMatches: noMatches},
		choicesOf(opened.listed),
	)
	opened.list = opened.list.WithCursorOn(offer.Current.String())

	return opened, cmd
}

func (p Picker) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg)
	case look.Resized:
		return outcome.Stay(p.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(p.restyled(msg.Styles))
	}

	return p.listUpdated(msg)
}

func (p Picker) View() string {
	return look.Frame(p.styles.Focused, p.offer.Title, p.list.View(p.styles), p.outer)
}

func (p Picker) ShortHelp() []key.Binding {
	return p.list.ShortHelp()
}

func (p Picker) FullHelp() [][]key.Binding {
	return p.list.FullHelp()
}

func (p Picker) pressed(msg tea.KeyPressMsg) outcome.Step {
	if len(p.offer.Curated) > 0 && p.keys.Matches(msg, binding.ShowAllLanguages) {
		return outcome.Stay(p.toggled())
	}

	return p.listUpdated(msg)
}

func (p Picker) listUpdated(msg tea.Msg) outcome.Step {
	next := p

	var (
		result picker.Result
		cmd    tea.Cmd
	)

	next.list, result, cmd = p.list.Update(msg)

	switch result.Ending {
	case picker.Picked:
		return outcome.Close().Passing(p.offer.Picked(p.listed[result.Index])).Running(cmd)
	case picker.Cancelled:
		return outcome.Close().Running(cmd)
	case picker.Filtering:
	}

	return outcome.Stay(next).Running(cmd)
}

func (p Picker) toggled() Picker {
	highlighted, _ := p.list.Highlighted()

	next := p
	next.showingAll = !p.showingAll
	next.listed = next.offered()
	next.list = p.list.WithChoices(choicesOf(next.listed)).WithCursorOn(highlighted.Text)

	return next
}

func (p Picker) resized(screen look.Size) Picker {
	next := p
	next.outer = screen.Share(pickerPercent)
	next.list, _, _ = p.list.Update(look.Resized{Box: next.outer.Inner()})

	return next
}

func (p Picker) restyled(styles look.Styles) Picker {
	p.styles = styles

	return p
}

func (p Picker) offered() []value.Language {
	if p.showingAll {
		return slices.SortedFunc(slices.Values(value.Languages()), compareIgnoringCase)
	}

	return p.offer.Curated
}

func compareIgnoringCase(a, b value.Language) int {
	return cmp.Or(
		strings.Compare(strings.ToLower(a.String()), strings.ToLower(b.String())),
		strings.Compare(a.String(), b.String()),
	)
}

func choicesOf(languages []value.Language) []picker.Choice {
	choices := make([]picker.Choice, 0, len(languages))
	for _, language := range languages {
		choices = append(choices, picker.Choice{Text: language.String(), Meta: ""})
	}

	return choices
}
