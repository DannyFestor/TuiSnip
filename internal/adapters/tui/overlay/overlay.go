package overlay

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Overlay interface {
	Update(msg tea.Msg) Step
	View() string
	Hints() []key.Binding
}
