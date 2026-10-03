package tui

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	searchTitle          = "Search"
	previewTitle         = "Preview"
	searchPrompt         = "/ "
	noMatchesText        = "No Snippets match."
	searchPopupPercent   = 80
	resultsListPercent   = 40
	queryRows            = 2
	resultWord           = " result"
	resultsWord          = " results"
	singleResult         = 1
	searchTitleSeparator = " · "
)

type popupHalves struct {
	results look.Size
	preview look.Size
}

type searchPopup struct {
	keys    searchBindings
	styles  look.Styles
	query   textinput.Model
	browse  []domain.Snippet
	results snippetList
	preview snippetPane
	outer   look.Size
}

func newSearchPopup(
	keys searchBindings,
	styles look.Styles,
	preview snippetPane,
	browse []domain.Snippet,
) (searchPopup, tea.Cmd) {
	popup := searchPopup{
		keys:    keys,
		styles:  styles,
		query:   input.NewLine(searchPrompt),
		browse:  browse,
		results: newSnippetList(folderPathOf).withSnippets(browse),
		preview: preview,
		outer:   look.Size{Width: 0, Height: 0},
	}
	cmd := popup.query.Focus()

	return popup.previewed(), cmd
}

func (p searchPopup) Update(msg tea.Msg) step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg)
	case tea.PasteMsg:
		return p.typed(msg)
	case look.Resized:
		return stay(p.resized(msg.Box))
	case tea.BackgroundColorMsg:
		return stay(p.withCodeStyle(look.CodeStyleFor(msg)))
	case searchFinishedMsg:
		return stay(p.withHits(msg.text, msg.hits))
	}

	return stay(p)
}

func (p searchPopup) View() string {
	halves := p.halves()
	rows := p.resultRows(halves.results.Inner().Width)
	resultsFrame := look.Frame(p.styles.Focused, p.frameTitle(), p.query.View()+"\n\n"+rows, halves.results)
	previewFrame := look.Frame(p.styles.Unfocused, previewTitle, p.preview.body(p.styles, nil), halves.preview)

	return lipgloss.JoinHorizontal(lipgloss.Top, resultsFrame, previewFrame)
}

func (p searchPopup) ShortHelp() []key.Binding {
	return p.keys.hints()
}

func (p searchPopup) FullHelp() [][]key.Binding {
	return [][]key.Binding{p.ShortHelp()}
}

func (p searchPopup) pressed(msg tea.KeyPressMsg) step {
	switch {
	case key.Matches(msg, p.keys.down):
		return stay(p.moved(move.Down))
	case key.Matches(msg, p.keys.up):
		return stay(p.moved(move.Up))
	case key.Matches(msg, p.keys.accept):
		return p.onSelected(revealing)
	case key.Matches(msg, p.keys.copy):
		return p.onSelected(copying)
	case key.Matches(msg, p.keys.cancel):
		return closing()
	}

	return p.typed(msg)
}

func (p searchPopup) onSelected(stepFor func(domain.SnippetID) step) step {
	selected, ok := p.results.selected()
	if !ok {
		return stay(p)
	}

	return stepFor(selected.ID())
}

func (p searchPopup) text() string {
	return p.query.Value()
}

func (p searchPopup) withHits(text string, hits []domain.SearchHit) searchPopup {
	if text != p.text() {
		return p
	}

	snippets := make([]domain.Snippet, 0, len(hits))
	for index := range hits {
		snippets = append(snippets, hits[index].Snippet())
	}

	return p.listing(snippets)
}

func (p searchPopup) withCodeStyle(codeStyle string) searchPopup {
	p.preview = p.preview.withCodeStyle(codeStyle)

	return p
}

func (p searchPopup) resized(screen look.Size) searchPopup {
	p.outer = screen.Share(searchPopupPercent)
	halves := p.halves()
	p.query.SetWidth(max(1, halves.results.Inner().Width-len(searchPrompt)-cursorCell))
	p.results = p.results.resized(max(0, halves.results.Inner().Height-queryRows))
	p.preview = p.preview.resized(halves.preview.Inner())

	return p
}

func (p searchPopup) typed(msg tea.Msg) step {
	next := p

	var cmd tea.Cmd

	next.query, cmd = p.query.Update(msg)
	if next.text() == p.text() {
		return stay(next).Running(cmd)
	}

	if value.NewSearchQuery(next.text()).IsBlank() {
		return stay(next.listing(next.browse)).Running(cmd)
	}

	return stay(next).Passing(outcome.SearchTyped{Text: next.text()}).Running(cmd)
}

func (p searchPopup) listing(snippets []domain.Snippet) searchPopup {
	next := p
	next.results = p.results.withSnippets(snippets).moved(move.Top)

	return next.previewed()
}

func (p searchPopup) moved(direction move.Direction) searchPopup {
	next := p
	next.results = p.results.moved(direction)

	return next.previewed()
}

func (p searchPopup) previewed() searchPopup {
	selected, ok := p.results.selected()
	if ok {
		p.preview = p.preview.showing(selected)
	} else {
		p.preview = p.preview.cleared()
	}

	return p
}

func (p searchPopup) halves() popupHalves {
	resultsWidth := p.outer.Width * resultsListPercent / look.Percent

	return popupHalves{
		results: look.Size{Width: resultsWidth, Height: p.outer.Height},
		preview: look.Size{Width: p.outer.Width - resultsWidth, Height: p.outer.Height},
	}
}

func (p searchPopup) resultRows(width int) string {
	if len(p.results.snippets) == 0 {
		return p.styles.Dim.Render(noMatchesText)
	}

	return p.results.rows(p.styles.Focused, width)
}

func (p searchPopup) frameTitle() string {
	count := len(p.results.snippets)
	noun := resultsWord

	if count == singleResult {
		noun = resultWord
	}

	return searchTitle + searchTitleSeparator + strconv.Itoa(count) + noun
}

func revealing(id domain.SnippetID) step {
	return closing().Passing(outcome.SnippetRevealed{ID: id})
}

func copying(id domain.SnippetID) step {
	return closing().Passing(outcome.CopyRequested{ID: id})
}
