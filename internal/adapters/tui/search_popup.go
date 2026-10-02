package tui

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	searchTitle          = "Search"
	previewTitle         = "Preview"
	searchPrompt         = "/ "
	noMatchesText        = "No Snippets match."
	resultsListPercent   = 40
	queryRows            = 2
	resultWord           = " result"
	resultsWord          = " results"
	singleResult         = 1
	searchTitleSeparator = " · "
)

type popupHalves struct {
	results size
	preview size
}

type searchPopup struct {
	open    bool
	keys    searchBindings
	query   textinput.Model
	browse  []domain.Snippet
	results snippetList
	preview snippetPane
	outer   size
}

func newSearchPopup(keys searchBindings, preview snippetPane, browse []domain.Snippet) (searchPopup, tea.Cmd) {
	popup := searchPopup{
		open:    true,
		keys:    keys,
		query:   newLineInput(searchPrompt),
		browse:  browse,
		results: newSnippetList(folderPathOf).withSnippets(browse),
		preview: preview,
		outer:   size{width: 0, height: 0},
	}
	cmd := popup.query.Focus()

	return popup.previewed(), cmd
}

func noSearchPopup() searchPopup {
	return searchPopup{}
}

func (p searchPopup) update(msg tea.Msg) (searchPopup, searchRequest, tea.Cmd) {
	pressed, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p.typed(msg)
	}

	switch {
	case key.Matches(pressed, p.keys.down):
		return p.moved(moveDown), searchStays, nil
	case key.Matches(pressed, p.keys.up):
		return p.moved(moveUp), searchStays, nil
	case key.Matches(pressed, p.keys.accept):
		return p, searchReveals, nil
	case key.Matches(pressed, p.keys.copy):
		return p, searchCopies, nil
	case key.Matches(pressed, p.keys.cancel):
		return p, searchCloses, nil
	}

	return p.typed(pressed)
}

func (p searchPopup) text() string {
	return p.query.Value()
}

func (p searchPopup) selected() (domain.Snippet, bool) {
	return p.results.selected()
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

func (p searchPopup) resized(screen size) searchPopup {
	p.outer = shareOf(screen, searchPopupPercent)
	halves := p.halves()
	p.query.SetWidth(max(1, innerSize(halves.results).width-len(searchPrompt)-cursorCell))
	p.results = p.results.resized(max(0, innerSize(halves.results).height-queryRows))
	p.preview = p.preview.resized(innerSize(halves.preview))

	return p
}

func (p searchPopup) view(styles styleSet) string {
	halves := p.halves()
	rows := p.resultRows(styles, innerSize(halves.results).width)
	resultsFrame := frame(styles.focused, p.frameTitle(), p.query.View()+"\n\n"+rows, halves.results)
	previewFrame := frame(styles.unfocused, previewTitle, p.preview.body(styles, nil), halves.preview)

	return lipgloss.JoinHorizontal(lipgloss.Top, resultsFrame, previewFrame)
}

func (p searchPopup) typed(msg tea.Msg) (searchPopup, searchRequest, tea.Cmd) {
	before := p.text()

	var cmd tea.Cmd

	p.query, cmd = p.query.Update(msg)
	if p.text() == before {
		return p, searchStays, cmd
	}

	if value.NewSearchQuery(p.text()).IsBlank() {
		return p.listing(p.browse), searchStays, cmd
	}

	return p, searchQueries, cmd
}

func (p searchPopup) listing(snippets []domain.Snippet) searchPopup {
	next := p
	next.results = p.results.withSnippets(snippets).moved(moveTop)

	return next.previewed()
}

func (p searchPopup) moved(move movement) searchPopup {
	next := p
	next.results = p.results.moved(move)

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
	resultsWidth := p.outer.width * resultsListPercent / percent

	return popupHalves{
		results: size{width: resultsWidth, height: p.outer.height},
		preview: size{width: p.outer.width - resultsWidth, height: p.outer.height},
	}
}

func (p searchPopup) resultRows(styles styleSet, width int) string {
	if len(p.results.snippets) == 0 {
		return styles.dim.Render(noMatchesText)
	}

	return p.results.rows(styles.focused, width)
}

func (p searchPopup) frameTitle() string {
	count := len(p.results.snippets)
	noun := resultsWord

	if count == singleResult {
		noun = resultWord
	}

	return searchTitle + searchTitleSeparator + strconv.Itoa(count) + noun
}

func (p searchPopup) hints() []key.Binding {
	return p.keys.hints()
}
