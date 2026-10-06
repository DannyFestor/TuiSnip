package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

type EditedContentHandler interface {
	Handle(edited outcome.ContentEdited, overlays outcome.Stack) (outcome.Stack, []outcome.Outcome, tea.Cmd)
}

type IntoEditOverlay struct{}

func (IntoEditOverlay) Handle(
	edited outcome.ContentEdited, overlays outcome.Stack,
) (outcome.Stack, []outcome.Outcome, tea.Cmd) {
	return overlays.Offered(edited)
}
