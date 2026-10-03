package outcome

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"

type Step = overlay.Step[Outcome]

func Stay(next Overlay) Step {
	return overlay.Stay(next)
}

func Close() Step {
	return overlay.Close[Outcome]()
}
