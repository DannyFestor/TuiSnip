package overlay

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const centring = 2

func centredOver(background string, screen tea.WindowSizeMsg, foreground string) string {
	x := max(0, (screen.Width-lipgloss.Width(foreground))/centring)
	y := max(0, (screen.Height-lipgloss.Height(foreground))/centring)

	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(background),
		lipgloss.NewLayer(foreground).X(x).Y(y).Z(1),
	)

	return lipgloss.NewCanvas(screen.Width, screen.Height).Compose(layers).Render()
}
