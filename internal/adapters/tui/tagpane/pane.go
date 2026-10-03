package tagpane

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

const noTagsText = "No Tags yet."

type Pane struct {
	keys   binding.Set
	styles look.Styles
	box    look.Size
}

func New(keys binding.Keys, styles look.Styles) Pane {
	return Pane{
		keys:   keys.For(binding.ScopeTags),
		styles: styles,
		box:    look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	if resized, ok := msg.(look.Resized); ok {
		p.box = resized.Box
	}

	return p, nil, nil
}

func (p Pane) View() string {
	return look.FitWidth(p.styles.Dim.Render(noTagsText), p.box.Width)
}

func (p Pane) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}
