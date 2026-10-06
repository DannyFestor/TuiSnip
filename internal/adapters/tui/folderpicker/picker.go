package folderpicker

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
)

const (
	pickerPercent = 60
	filterPrompt  = "filter: "
	noMatches     = "No Folder matches."
)

type Picker struct {
	styles       look.Styles
	offer        Offer
	destinations []destination
	list         picker.List
	outer        look.Size
}

func New(keys binding.Keys, styles look.Styles, offer Offer) (Picker, tea.Cmd) {
	destinations := destinationsIn(offer.Tree, offer.Moving)
	list, cmd := picker.New(
		keys,
		picker.Labels{Prompt: filterPrompt, NoMatches: noMatches},
		choicesOf(destinations, folderpath.New(offer.Tree)),
	)

	return Picker{
		styles:       styles,
		offer:        offer,
		destinations: destinations,
		list: list.WithGreyedOut(refusedIndexes(destinations)...).
			WithCursorOnChoice(indexOf(destinations, offer.Current)),
		outer: look.Size{Width: 0, Height: 0},
	}, cmd
}

func (p Picker) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
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

func (p Picker) listUpdated(msg tea.Msg) outcome.Step {
	next := p

	var (
		result picker.Result
		cmd    tea.Cmd
	)

	next.list, result, cmd = p.list.Update(msg)

	switch result.Ending {
	case picker.Picked:
		return outcome.Close().Passing(p.offer.Picked(p.destinations[result.Index].folderID)).Running(cmd)
	case picker.Cancelled:
		return outcome.Close().Running(cmd)
	case picker.Filtering:
	}

	return outcome.Stay(next).Running(cmd)
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
