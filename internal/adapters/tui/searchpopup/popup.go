package searchpopup

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	searchTitle        = "Search"
	previewTitle       = "Preview"
	searchPrompt       = "/ "
	noMatchesText      = "No Snippets match."
	popupPercent       = 80
	resultsListPercent = 40
	queryRows          = 2
	resultWord         = " result"
	resultsWord        = " results"
	singleResult       = 1
	titleSeparator     = " · "
)

type halves struct {
	results look.Size
	preview look.Size
}

type Popup struct {
	keys    binding.Set
	styles  look.Styles
	query   textinput.Model
	browse  []domain.Snippet
	results snippetlist.List
	preview snippetpane.Pane
	outer   look.Size
}

func New(keys binding.Keys, styles look.Styles, preview snippetpane.Pane, listing Listing) (Popup, tea.Cmd) {
	popup := Popup{
		keys:    keys.For(binding.ScopeSearch),
		styles:  styles,
		query:   input.NewLine(searchPrompt),
		browse:  listing.Snippets,
		results: snippetlist.New(keys, styles, shortPathIn(listing.Paths)).WithSnippets(listing.Snippets),
		preview: preview.WithPaths(listing.Paths),
		outer:   look.Size{Width: 0, Height: 0},
	}
	cmd := popup.query.Focus()

	return popup.previewed(), cmd
}

func (p Popup) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg)
	case tea.PasteMsg:
		return p.typed(msg)
	case look.Resized:
		return outcome.Stay(p.resized(msg.Box))
	case look.Restyled:
		return outcome.Stay(p.restyled(msg))
	case HitsFound:
		return outcome.Stay(p.withHits(msg.Text, msg.Hits))
	case arrived.Tree:
		return outcome.Stay(p.withPaths(folderpath.New(msg.Tree)))
	}

	return outcome.Stay(p)
}

func (p Popup) View() string {
	split := p.halves()
	rows := p.resultRows()
	resultsFrame := look.Frame(p.styles.Focused, p.frameTitle(), p.query.View()+"\n\n"+rows, split.results)
	previewFrame := look.Frame(p.styles.Unfocused, previewTitle, p.previewBody(), split.preview)

	return lipgloss.JoinHorizontal(lipgloss.Top, resultsFrame, previewFrame)
}

func (p Popup) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Popup) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}

func (p Popup) pressed(msg tea.KeyPressMsg) outcome.Step {
	switch {
	case p.keys.Matches(msg, binding.Down):
		return outcome.Stay(p.moved(move.Down))
	case p.keys.Matches(msg, binding.Up):
		return outcome.Stay(p.moved(move.Up))
	case p.keys.Matches(msg, binding.Accept):
		return p.onSelected(revealing)
	case p.keys.Matches(msg, binding.Copy):
		return p.onSelected(copying)
	case p.keys.Matches(msg, binding.Cancel):
		return outcome.Close()
	}

	return p.typed(msg)
}

func (p Popup) onSelected(stepFor func(domain.Snippet) outcome.Step) outcome.Step {
	selected, ok := p.results.Selected()
	if !ok {
		return outcome.Stay(p)
	}

	return stepFor(selected)
}

func (p Popup) text() string {
	return p.query.Value()
}

func (p Popup) withHits(text string, hits []domain.SearchHit) Popup {
	if text != p.text() {
		return p
	}

	snippets := make([]domain.Snippet, 0, len(hits))
	for index := range hits {
		snippets = append(snippets, hits[index].Snippet())
	}

	return p.listing(snippets)
}

func (p Popup) withPaths(paths folderpath.Paths) Popup {
	p.results = p.results.WithMeta(shortPathIn(paths))
	p.preview = p.preview.WithPaths(paths)

	return p
}

func (p Popup) restyled(msg look.Restyled) Popup {
	p.styles = msg.Styles
	p.results, _, _ = p.results.Update(msg)
	p.preview, _, _ = p.preview.Update(msg)

	return p
}

func (p Popup) resized(screen look.Size) Popup {
	p.outer = screen.Share(popupPercent)
	split := p.halves()
	p.query.SetWidth(max(1, split.results.Inner().Width-len(searchPrompt)-input.CursorWidth))
	resultsInner := split.results.Inner()
	resultRows := look.Size{Width: resultsInner.Width, Height: max(0, resultsInner.Height-queryRows)}
	p.results, _, _ = p.results.Update(look.Resized{Box: resultRows})
	p.preview, _, _ = p.preview.Update(look.Resized{Box: split.preview.Inner()})

	return p
}

func (p Popup) typed(msg tea.Msg) outcome.Step {
	next := p

	var cmd tea.Cmd

	next.query, cmd = p.query.Update(msg)
	if next.text() == p.text() {
		return outcome.Stay(next).Running(cmd)
	}

	if value.NewSearchQuery(next.text()).IsBlank() {
		return outcome.Stay(next.listing(next.browse)).Running(cmd)
	}

	return outcome.Stay(next).Passing(outcome.SearchTyped{Text: next.text()}).Running(cmd)
}

func (p Popup) listing(snippets []domain.Snippet) Popup {
	next := p
	next.results = p.results.WithSnippets(snippets).Moved(move.Top)

	return next.previewed()
}

func (p Popup) moved(direction move.Direction) Popup {
	next := p
	next.results = p.results.Moved(direction)

	return next.previewed()
}

func (p Popup) previewed() Popup {
	selected, ok := p.results.Selected()
	if ok {
		p.preview = p.preview.Showing(selected)
	} else {
		p.preview = p.preview.Cleared()
	}

	return p
}

func (p Popup) halves() halves {
	resultsWidth := p.outer.Width * resultsListPercent / look.Percent

	return halves{
		results: look.Size{Width: resultsWidth, Height: p.outer.Height},
		preview: look.Size{Width: p.outer.Width - resultsWidth, Height: p.outer.Height},
	}
}

func (p Popup) resultRows() string {
	if p.noResults() {
		return p.styles.Dim.Render(noMatchesText)
	}

	return p.results.View(p.styles.Focused)
}

func (p Popup) previewBody() string {
	if p.noResults() {
		return look.EmptyHint(p.styles, nil)
	}

	return p.preview.View()
}

func (p Popup) noResults() bool {
	return len(p.results.Snippets()) == 0
}

func (p Popup) frameTitle() string {
	count := len(p.results.Snippets())
	noun := resultsWord

	if count == singleResult {
		noun = resultWord
	}

	return searchTitle + titleSeparator + strconv.Itoa(count) + noun
}

func revealing(selected domain.Snippet) outcome.Step {
	return outcome.Close().Passing(outcome.SnippetRevealed{ID: selected.ID(), FolderID: selected.FolderID()})
}

func copying(selected domain.Snippet) outcome.Step {
	return outcome.Close().Passing(outcome.CopyRequested{ID: selected.ID()})
}
