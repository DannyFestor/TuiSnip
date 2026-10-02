package overlay

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Stack struct {
	layers []Overlay
	screen tea.WindowSizeMsg
}

func NewStack() Stack {
	return Stack{layers: nil, screen: tea.WindowSizeMsg{Width: 0, Height: 0}}
}

func (s Stack) Open() bool {
	return len(s.layers) > 0
}

func (s Stack) Hints() []key.Binding {
	if !s.Open() {
		return nil
	}

	return s.layers[len(s.layers)-1].Hints()
}

func (s Stack) Pushed(pushed Overlay) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)
	current.opened(len(s.layers)-1, pushed)

	return current.finished()
}

func (s Stack) Update(msg tea.Msg) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)

	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		current.routedToTop(msg)
	case tea.WindowSizeMsg:
		current.screen = msg
		current.delivered(msg)
	default:
		current.delivered(msg)
	}

	return current.finished()
}

func (s Stack) Offered(outcome Outcome) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)
	current.bubbled(len(s.layers)-1, outcome)

	return current.finished()
}

func (s Stack) Render(background string) string {
	view := background
	for _, layer := range s.layers {
		view = centredOver(view, s.screen, layer.View())
	}

	return view
}
