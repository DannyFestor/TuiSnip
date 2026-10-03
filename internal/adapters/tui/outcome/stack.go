package outcome

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"

type Stack = overlay.Stack[Outcome]

func NewStack() Stack {
	return overlay.NewStack[Outcome]()
}
