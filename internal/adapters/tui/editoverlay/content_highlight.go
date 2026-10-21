package editoverlay

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

// Lexing 256 KiB of Content takes about 250 ms, too long for every key, so
// the Content is lexed off the UI thread once typing pauses this long.
const highlightPause = 150 * time.Millisecond

type contentHighlight struct {
	scheduled int
	shown     int
	requested highlightSource
}

type highlightSource struct {
	content      string
	language     value.Language
	codeStyle    string
	colourResets int
}

func (h contentHighlight) rescheduled() (contentHighlight, tea.Cmd) {
	h.scheduled++
	generation := h.scheduled

	return h, tea.Tick(highlightPause, func(time.Time) tea.Msg { return HighlightDue{generation: generation} })
}

func (h contentHighlight) due(msg HighlightDue, source highlightSource) (contentHighlight, tea.Cmd) {
	if msg.generation != h.scheduled || source == h.requested {
		return h, nil
	}

	h.requested = source

	return h, func() tea.Msg {
		return ContentHighlighted{
			generation:   msg.generation,
			colourResets: source.colourResets,
			colours:      source.colours(),
		}
	}
}

func (h contentHighlight) landed(msg ContentHighlighted, colourResets int) (contentHighlight, bool) {
	if msg.generation <= h.shown || msg.colourResets != colourResets {
		return h, false
	}

	h.shown = msg.generation

	return h, true
}

func (s highlightSource) colours() look.LineColours {
	if s.language == value.PlainText() {
		return look.LineColours{}
	}

	return look.ColourLines(s.content, s.language.String(), s.codeStyle)
}
