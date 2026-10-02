package overlay

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Stack struct {
	overlays []Overlay
	screen   tea.WindowSizeMsg
}

func NewStack() Stack {
	return Stack{overlays: nil, screen: tea.WindowSizeMsg{Width: 0, Height: 0}}
}

func (s Stack) Open() bool {
	return len(s.overlays) > 0
}

func (s Stack) Hints() []key.Binding {
	if !s.Open() {
		return nil
	}

	return s.overlays[len(s.overlays)-1].Hints()
}

func (s Stack) Pushed(pushed Overlay) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)
	current.open(len(s.overlays)-1, pushed)

	return current.finish()
}

func (s Stack) Update(msg tea.Msg) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)

	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		current.routeToTop(msg)
	case tea.WindowSizeMsg:
		current.screen = msg
		current.deliver(msg)
	default:
		current.deliver(msg)
	}

	return current.finish()
}

func (s Stack) Offered(outcome Outcome) (Stack, []Outcome, tea.Cmd) {
	current := roundOver(s)
	current.bubble(len(s.overlays)-1, outcome)

	return current.finish()
}

func (s Stack) Render(background string) string {
	view := background
	for _, drawn := range s.overlays {
		view = centredOver(view, s.screen, drawn.View())
	}

	return view
}
