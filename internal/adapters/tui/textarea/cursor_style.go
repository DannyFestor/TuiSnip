package textarea

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
)

type CursorStyle struct {
	Color      color.Color
	Shape      tea.CursorShape
	Blink      bool
	BlinkSpeed time.Duration
}
