package folderpicker

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
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
	steered      bool
}

func New(keys binding.Keys, styles look.Styles, offer Offer) (Picker, tea.Cmd) {
	list, cmd := picker.New(keys, picker.Labels{Prompt: filterPrompt, NoMatches: noMatches}, nil)
	opened := Picker{
		styles:       styles,
		offer:        offer,
		destinations: nil,
		list:         list,
		outer:        look.Size{Width: 0, Height: 0},
		steered:      false,
	}

	return opened.withTree(offer.Tree), cmd
}

func (p Picker) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case look.Resized:
		return outcome.Stay(p.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(p.restyled(msg.Styles))
	case arrived.Tree:
		return outcome.Stay(p.withTree(msg.Tree))
	case tea.KeyPressMsg:
		return p.steeredBy(msg)
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

func (p Picker) withTree(tree browse.Tree) Picker {
	landing, lands := p.landing()
	next := p
	next.destinations = destinationsIn(tree, p.offer.Moving)
	next.list = p.list.WithChoices(choicesOf(next.destinations, folderpath.New(tree))).
		WithGreyedOut(refusedIndexes(next.destinations)...)

	if lands {
		next.list = next.list.WithCursorOnChoice(indexOf(next.destinations, landing))
	}

	return next
}

func (p Picker) landing() (domain.FolderID, bool) {
	if !p.steered {
		return p.offer.Current, true
	}

	index, ok := p.list.HighlightedIndex()
	if !ok {
		return domain.FolderID{}, false
	}

	return p.destinations[index].folderID, true
}

func (p Picker) steeredBy(msg tea.KeyPressMsg) outcome.Step {
	next := p
	next.steered = true

	return next.listUpdated(msg)
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
