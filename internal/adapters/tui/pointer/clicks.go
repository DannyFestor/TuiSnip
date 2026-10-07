package pointer

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Terminals report every press on its own, so a double click is two presses on one cell within this pause,
// the default double-click time on Windows and most Linux desktops.
const doubleClickPause = 500 * time.Millisecond

type Clicks struct {
	last    Point
	at      time.Time
	pending bool
}

func (c Clicks) Next(msg tea.MouseClickMsg, now time.Time) (Clicks, Clicked, bool) {
	if msg.Button != tea.MouseLeft {
		return c, Clicked{At: Point{X: 0, Y: 0}, Double: false}, false
	}

	at := Point{X: msg.X, Y: msg.Y}
	double := c.pending && at == c.last && now.Sub(c.at) <= doubleClickPause

	return Clicks{last: at, at: now, pending: !double}, Clicked{At: at, Double: double}, true
}
