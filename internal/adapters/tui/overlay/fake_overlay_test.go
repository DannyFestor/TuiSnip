package overlay_test

import (
	"fmt"
	"maps"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

type heard struct {
	by   string
	what string
}

type relayed struct {
	by      string
	outcome overlay.Outcome
}

type tickMsg struct{}

type fakeOverlay struct {
	name     string
	look     string
	reacts   map[string]func(fakeOverlay) overlay.Step
	receives func(fakeOverlay, overlay.Outcome) overlay.Step
}

type childlessOverlay struct{}

func (c childlessOverlay) Update(tea.Msg) overlay.Step {
	return overlay.Stay(c)
}

func (childlessOverlay) View() string {
	return "childless"
}

func (childlessOverlay) Hints() []key.Binding {
	return nil
}

func newFake(name string) fakeOverlay {
	return fakeOverlay{name: name, look: name, reacts: map[string]func(fakeOverlay) overlay.Step{}, receives: nil}
}

func (f fakeOverlay) Update(msg tea.Msg) overlay.Step {
	what := describe(msg)
	if react, ok := f.reacts[what]; ok {
		return react(f)
	}

	return overlay.Stay(f).Passing(heard{by: f.name, what: what})
}

func (f fakeOverlay) Received(outcome overlay.Outcome) overlay.Step {
	if f.receives != nil {
		return f.receives(f, outcome)
	}

	return overlay.Stay(f).Passing(outcome)
}

func (f fakeOverlay) View() string {
	return f.look
}

func (f fakeOverlay) Hints() []key.Binding {
	return []key.Binding{key.NewBinding(key.WithKeys(f.name), key.WithHelp(f.name, f.name))}
}

func (f fakeOverlay) looking(look string) fakeOverlay {
	f.look = look

	return f
}

func (f fakeOverlay) on(text string, react func(fakeOverlay) overlay.Step) fakeOverlay {
	f.reacts = maps.Clone(f.reacts)
	f.reacts[text] = react

	return f
}

func (f fakeOverlay) receiving(receive func(fakeOverlay, overlay.Outcome) overlay.Step) fakeOverlay {
	f.receives = receive

	return f
}

func closing(fakeOverlay) overlay.Step {
	return overlay.Close()
}

func staying(f fakeOverlay, _ overlay.Outcome) overlay.Step {
	return overlay.Stay(f)
}

func closingOn(want overlay.Outcome) func(fakeOverlay, overlay.Outcome) overlay.Step {
	return func(f fakeOverlay, outcome overlay.Outcome) overlay.Step {
		if outcome != want {
			return overlay.Stay(f)
		}

		return overlay.Close()
	}
}

func relaying(f fakeOverlay, outcome overlay.Outcome) overlay.Step {
	return overlay.Stay(f).Passing(relayed{by: f.name, outcome: outcome})
}

func opening(child fakeOverlay) func(fakeOverlay) overlay.Step {
	return func(f fakeOverlay) overlay.Step {
		return overlay.Stay(f).Opening(child)
	}
}

func describe(msg tea.Msg) string {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return msg.String()
	case tea.PasteMsg:
		return "paste " + msg.Content
	case tea.WindowSizeMsg:
		return fmt.Sprintf("%dx%d", msg.Width, msg.Height)
	case tickMsg:
		return "tick"
	}

	return fmt.Sprintf("%T", msg)
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func hintKeys(stack overlay.Stack) []string {
	keys := make([]string, 0, len(stack.Hints()))
	for _, binding := range stack.Hints() {
		keys = append(keys, binding.Keys()...)
	}

	return keys
}

func stackOf(overlays ...overlay.Overlay) overlay.Stack {
	stack := overlay.NewStack()
	for _, pushed := range overlays {
		stack, _, _ = stack.Pushed(pushed)
	}

	return stack
}
