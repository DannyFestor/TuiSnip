package move

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
)

type movementBinding struct {
	name      string
	direction Direction
}

func Pressed(global binding.Set, msg tea.KeyPressMsg) (Direction, bool) {
	for _, candidate := range movementBindings() {
		if global.Matches(msg, candidate.name) {
			return candidate.direction, true
		}
	}

	return None, false
}

func movementBindings() []movementBinding {
	return []movementBinding{
		{name: binding.Down, direction: Down},
		{name: binding.Up, direction: Up},
		{name: binding.Top, direction: Top},
		{name: binding.Bottom, direction: Bottom},
		{name: binding.PageDown, direction: PageDown},
		{name: binding.PageUp, direction: PageUp},
	}
}
