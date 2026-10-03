package overlay

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

//nolint:iface // iface misses calls through an instantiated generic interface; round calls every method.
type Overlay[O any] interface {
	help.KeyMap
	Update(msg tea.Msg) Step[O]
	View() string
}
