package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

type mouse struct {
	on     bool
	clock  Clock
	clicks pointer.Clicks
}

func (m mouse) mode() tea.MouseMode {
	if !m.on {
		return tea.MouseModeNone
	}

	return tea.MouseModeCellMotion
}

func (m mouse) clicked(msg tea.MouseClickMsg) (mouse, pointer.Clicked, bool) {
	if !m.on {
		return m, pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false}, false
	}

	next := m

	var (
		click pointer.Clicked
		ok    bool
	)

	next.clicks, click, ok = m.clicks.Next(msg, m.clock.Now())

	return next, click, ok
}

func (m mouse) wheeled(msg tea.MouseWheelMsg) (pointer.Wheeled, bool) {
	wheel, ok := pointer.WheeledBy(msg)

	return wheel, ok && m.on
}
