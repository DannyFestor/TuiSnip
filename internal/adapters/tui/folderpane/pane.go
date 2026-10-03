package folderpane

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

const rootLabel = "◆ Root"

type Pane struct {
	keys             binding.Set
	rootSnippetCount int
	box              look.Size
}

func New(keys binding.Keys) Pane {
	return Pane{
		keys:             keys.For(binding.ScopeFolders),
		rootSnippetCount: 0,
		box:              look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	if resized, ok := msg.(look.Resized); ok {
		p.box = resized.Box
	}

	return p, nil, nil
}

func (p Pane) View(frame look.FrameStyle) string {
	return frame.Cursor.Render(look.Row(rootLabel, strconv.Itoa(p.rootSnippetCount), p.box.Width))
}

func (p Pane) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}

func (p Pane) WithRootSnippetCount(count int) Pane {
	p.rootSnippetCount = count

	return p
}
