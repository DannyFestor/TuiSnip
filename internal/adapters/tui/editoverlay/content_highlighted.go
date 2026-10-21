package editoverlay

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"

type ContentHighlighted struct {
	generation   int
	colourResets int
	colours      look.LineColours
}
