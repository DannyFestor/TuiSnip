package trashlist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/cursorlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	folderMarker = "▸ "
	snippetNoun  = "Snippet"
)

type List struct {
	keys            binding.Set
	styles          look.Styles
	snippetLanguage func(domain.Snippet) string
	items           cursorlist.List[trash.Item]
}

func New(keys binding.Keys, styles look.Styles, snippetLanguage func(domain.Snippet) string) List {
	return List{
		keys:            keys.For(binding.ScopeTrash),
		styles:          styles,
		snippetLanguage: snippetLanguage,
		items:           cursorlist.New[trash.Item](keys),
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (l List) Update(msg tea.Msg) (List, []outcome.Outcome, tea.Cmd) {
	if restyled, ok := msg.(look.Restyled); ok {
		l.styles = restyled.Styles

		return l, nil, nil
	}

	l.items = l.items.Update(msg)

	return l, nil, nil
}

func (l List) View(frame look.FrameStyle) string {
	if len(l.items.Rows()) == 0 {
		return look.EmptyTrash(l.styles)
	}

	return l.items.View(l.row, frame)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithItems(items []trash.Item) List {
	l.items = l.items.WithRows(items)

	return l
}

func (l List) Selected() (trash.Item, bool) {
	return l.items.Selected()
}

func (l List) HasRowAt(at pointer.Point) bool {
	return l.items.HasRowAt(at)
}

func (l List) row(item trash.Item, width int) string {
	if item.Kind.IsFolder() {
		return look.Row(folderMarker+item.Name(), look.Counted(item.Contents.SnippetCount, snippetNoun), width)
	}

	return look.Row(item.Name(), l.snippetLanguage(item.Snippet), width)
}
