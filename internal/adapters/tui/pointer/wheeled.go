package pointer

import tea "charm.land/bubbletea/v2"

// Bubbles' viewport scrolls this far per wheel turn, so every Pane does the same.
const linesPerTurn = 3

type Wheeled struct {
	At    Point
	Lines int
}

func WheeledBy(msg tea.MouseWheelMsg) (Wheeled, bool) {
	lines := linesTurnedBy(msg.Button)
	if lines == 0 {
		return Wheeled{At: Point{X: 0, Y: 0}, Lines: 0}, false
	}

	return Wheeled{At: Point{X: msg.X, Y: msg.Y}, Lines: lines}, true
}

func (w Wheeled) Relative(origin Point) Wheeled {
	w.At = w.At.Relative(origin)

	return w
}

func linesTurnedBy(button tea.MouseButton) int {
	if button == tea.MouseWheelDown {
		return linesPerTurn
	}

	if button == tea.MouseWheelUp {
		return -linesPerTurn
	}

	return 0
}
