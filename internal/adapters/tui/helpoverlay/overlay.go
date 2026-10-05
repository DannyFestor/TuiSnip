package helpoverlay

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

const title = "Help"

type Overlay struct {
	keys    binding.Set
	hints   []key.Binding
	listed  [][]key.Binding
	styles  look.Styles
	columns help.Model
	screen  look.Size
}

func New(keys binding.Keys, styles look.Styles, listed [][]key.Binding) Overlay {
	return Overlay{
		keys:    keys.For(binding.ScopeGlobal),
		hints:   keys.HelpOverlayHints(),
		listed:  listed,
		styles:  styles,
		columns: columnsStyled(styles),
		screen:  look.Size{Width: 0, Height: 0},
	}
}

func (o Overlay) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return o.pressed(msg)
	case look.Resized:
		return outcome.Stay(o.resized(msg.Box))
	}

	return outcome.Stay(o)
}

func (o Overlay) View() string {
	body := o.columns.FullHelpView(o.listed)
	outer := look.Size{
		Width:  min(lipgloss.Width(body)+look.BorderWidth, o.screen.Width),
		Height: min(lipgloss.Height(body)+look.BorderWidth, o.screen.Height),
	}

	return look.Frame(o.styles.Focused, title, body, outer)
}

func (o Overlay) ShortHelp() []key.Binding {
	return o.hints
}

func (o Overlay) FullHelp() [][]key.Binding {
	return [][]key.Binding{o.ShortHelp()}
}

func (o Overlay) pressed(msg tea.KeyPressMsg) outcome.Step {
	if o.keys.Matches(msg, binding.Help) || o.keys.Matches(msg, binding.Back) {
		return outcome.Close()
	}

	return outcome.Stay(o)
}

func (o Overlay) resized(screen look.Size) Overlay {
	o.screen = screen
	o.columns.SetWidth(screen.Inner().Width)

	return o
}

func columnsStyled(styles look.Styles) help.Model {
	columns := help.New()
	columns.Styles.FullKey = styles.Bold
	columns.Styles.FullDesc = styles.Plain
	columns.Styles.FullSeparator = styles.Plain

	return columns
}
