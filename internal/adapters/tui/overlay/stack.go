package overlay

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

type Stack[O any] struct {
	overlays []Overlay[O]
	screen   look.Size
}

func NewStack[O any]() Stack[O] {
	return Stack[O]{overlays: nil, screen: look.Size{Width: 0, Height: 0}}
}

func (s Stack[O]) Open() bool {
	return len(s.overlays) > 0
}

func (s Stack[O]) ShortHelp() []key.Binding {
	if !s.Open() {
		return nil
	}

	return s.top().ShortHelp()
}

func (s Stack[O]) FullHelp() [][]key.Binding {
	if !s.Open() {
		return nil
	}

	return s.top().FullHelp()
}

func (s Stack[O]) Pushed(pushed Overlay[O]) (Stack[O], []O, tea.Cmd) {
	current := roundOver(s)
	current.open(len(s.overlays)-1, pushed)

	return current.finish()
}

func (s Stack[O]) Update(msg tea.Msg) (Stack[O], []O, tea.Cmd) {
	current := roundOver(s)

	switch msg := msg.(type) {
	case tea.KeyPressMsg, tea.PasteMsg:
		current.routeToTop(msg)
	case pointer.Clicked:
		current.clickOnTop(msg)
	case pointer.Wheeled:
		current.wheelOnTop(msg)
	case tea.WindowSizeMsg:
		current.screen = look.SizeOf(msg)
		current.deliver(look.Resized{Box: current.screen})
	default:
		current.deliver(msg)
	}

	return current.finish()
}

func (s Stack[O]) Offered(outcome O) (Stack[O], []O, tea.Cmd) {
	current := roundOver(s)
	current.bubble(len(s.overlays)-1, outcome)

	return current.finish()
}

func (s Stack[O]) Render() string {
	view := ""
	for _, drawn := range s.overlays {
		view = s.drawnOver(view, drawn)
	}

	return view
}

func (s Stack[O]) drawnOver(view string, drawn Overlay[O]) string {
	if base, ok := drawn.(Base[O]); ok {
		return base.ViewUnder(s.ShortHelp())
	}

	return centredOver(view, s.screen, drawn.View())
}

func (s Stack[O]) top() Overlay[O] {
	return s.overlays[len(s.overlays)-1]
}
