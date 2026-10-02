package tui

import "charm.land/lipgloss/v2"

const (
	editOverlayPercent = 90
	searchPopupPercent = 80
	overlayLayers      = 2
	centring           = 2
)

func overlaid(background string, screen size, foreground string) string {
	x := max(0, (screen.width-lipgloss.Width(foreground))/centring)
	y := max(0, (screen.height-lipgloss.Height(foreground))/centring)

	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(background),
		lipgloss.NewLayer(foreground).X(x).Y(y).Z(1),
	)

	return lipgloss.NewCanvas(screen.width, screen.height).Compose(layers).Render()
}

func shareOf(screen size, sharePercent int) size {
	return size{width: screen.width * sharePercent / percent, height: screen.height * sharePercent / percent}
}
