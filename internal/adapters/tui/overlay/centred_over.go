package overlay

import (
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

const centring = 2

type placement struct {
	origin pointer.Point
	size   look.Size
}

func centredIn(screen look.Size, foreground string) placement {
	size := look.Size{Width: lipgloss.Width(foreground), Height: lipgloss.Height(foreground)}
	origin := pointer.Point{
		X: max(0, (screen.Width-size.Width)/centring),
		Y: max(0, (screen.Height-size.Height)/centring),
	}

	return placement{origin: origin, size: size}
}

func wholeScreen(screen look.Size) placement {
	return placement{origin: pointer.Point{X: 0, Y: 0}, size: screen}
}

func (p placement) holds(at pointer.Point) bool {
	return at.Relative(p.origin).Within(p.size)
}

func centredOver(background string, screen look.Size, foreground string) string {
	placed := centredIn(screen, foreground)

	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(background),
		lipgloss.NewLayer(foreground).X(placed.origin.X).Y(placed.origin.Y).Z(1),
	)

	return lipgloss.NewCanvas(screen.Width, screen.Height).Compose(layers).Render()
}
