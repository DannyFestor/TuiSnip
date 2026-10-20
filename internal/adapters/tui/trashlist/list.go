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
	trashItems      cursorlist.List[trash.Item]
}

func New(keys binding.Keys, styles look.Styles, snippetLanguage func(domain.Snippet) string) List {
	return List{
		keys:            keys.For(binding.ScopeTrash),
		styles:          styles,
		snippetLanguage: snippetLanguage,
		trashItems:      cursorlist.New[trash.Item](keys),
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (l List) Update(msg tea.Msg) (List, []outcome.Outcome, tea.Cmd) {
	if restyled, ok := msg.(look.Restyled); ok {
		l.styles = restyled.Styles

		return l, nil, nil
	}

	l.trashItems = l.trashItems.Update(msg)

	return l, nil, nil
}

func (l List) View(frame look.FrameStyle) string {
	if len(l.trashItems.Rows()) == 0 {
		return look.EmptyTrash(l.styles)
	}

	return l.trashItems.View(l.row, frame)
}

func (l List) ShortHelp() []key.Binding {
	return l.keys.ShortHelp()
}

func (l List) FullHelp() [][]key.Binding {
	return l.keys.FullHelp()
}

func (l List) WithItems(trashItems []trash.Item) List {
	l.trashItems = l.trashItems.WithRows(trashItems)

	return l
}

func (l List) TrashItems() []trash.Item {
	return l.trashItems.Rows()
}

func (l List) TrashItemIDs() []trash.ItemID {
	trashItems := l.trashItems.Rows()
	ids := make([]trash.ItemID, 0, len(trashItems))

	for index := range trashItems {
		ids = append(ids, trashItems[index].ID())
	}

	return ids
}

func (l List) Selected() (trash.Item, bool) {
	return l.trashItems.Selected()
}

func (l List) HasRowAt(at pointer.Point) bool {
	return l.trashItems.HasRowAt(at)
}

func (l List) row(trashItem trash.Item, width int) string {
	if trashItem.Kind.IsFolder() {
		return look.Row(
			folderMarker+trashItem.Name(),
			look.Counted(trashItem.Contents.SnippetCount, snippetNoun),
			width,
		)
	}

	return look.Row(trashItem.Name(), l.snippetLanguage(trashItem.Snippet), width)
}
