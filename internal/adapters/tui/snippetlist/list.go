package snippetlist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/cursorlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type List struct {
	keys       binding.Set
	emptyHints []key.Binding
	styles     look.Styles
	meta       Meta
	snippets   cursorlist.List[domain.Snippet]
}

func New(keys binding.Keys, styles look.Styles, meta Meta) List {
	return List{
		keys:       keys.For(binding.ScopeSnippetList),
		emptyHints: keys.EmptyListHints(),
		styles:     styles,
		meta:       meta,
		snippets:   cursorlist.New[domain.Snippet](keys),
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (l List) Update(msg tea.Msg) (List, []outcome.Outcome, tea.Cmd) {
	if restyled, ok := msg.(look.Restyled); ok {
		l.styles = restyled.Styles

		return l, nil, nil
	}

	l.snippets = l.snippets.Update(msg)

	return l, nil, nil
}

func (l List) View(frame look.FrameStyle) string {
	if len(l.snippets.Rows()) == 0 {
		return look.EmptyHint(l.styles, l.emptyHints)
	}

	return l.snippets.View(l.row, frame)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithSnippets(snippets []domain.Snippet) List {
	l.snippets = l.snippets.WithRows(snippets)

	return l
}

func (l List) WithMeta(meta Meta) List {
	l.meta = meta

	return l
}

func (l List) WithCursorOn(id domain.SnippetID) List {
	l.snippets = l.snippets.WithCursorOn(func(candidate domain.Snippet) bool { return candidate.ID() == id })

	return l
}

func (l List) Moved(direction move.Direction) List {
	l.snippets = l.snippets.Moved(direction)

	return l
}

func (l List) Selected() (domain.Snippet, bool) {
	return l.snippets.Selected()
}

func (l List) Successor() (domain.Snippet, bool) {
	return l.snippets.Successor()
}

func (l List) Snippets() []domain.Snippet {
	return l.snippets.Rows()
}

func (l List) HasRowAt(at pointer.Point) bool {
	return l.snippets.HasRowAt(at)
}

func (l List) row(snippet domain.Snippet, width int) string {
	return look.Row(snippet.Title().String(), l.meta(snippet), width)
}
