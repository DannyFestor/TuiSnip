package tui

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

type (
	outcomeOverlay = overlay.Overlay[outcome.Outcome]
	overlayStack   = overlay.Stack[outcome.Outcome]
	step           = overlay.Step[outcome.Outcome]
)

func newOverlayStack() overlayStack {
	return overlay.NewStack[outcome.Outcome]()
}

func stay(next outcomeOverlay) step {
	return overlay.Stay(next)
}

func closing() step {
	return overlay.Close[outcome.Outcome]()
}
