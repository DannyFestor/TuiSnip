package overlay_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

type tickMsg struct{}

func overlayMock(t *testing.T) *MockOverlay {
	t.Helper()

	mocked := NewMockOverlay(t)
	mocked.EXPECT().Update(mock.AnythingOfType("tea.WindowSizeMsg")).Return(overlay.Stay(mocked)).Maybe()

	return mocked
}

func parentMock(t *testing.T) *MockParent {
	t.Helper()

	mocked := NewMockParent(t)
	mocked.EXPECT().Update(mock.AnythingOfType("tea.WindowSizeMsg")).Return(overlay.Stay(mocked)).Maybe()

	return mocked
}

func hinting(name string) []key.Binding {
	return []key.Binding{key.NewBinding(key.WithKeys(name), key.WithHelp(name, name))}
}

func hintKeys(stack overlay.Stack) []string {
	keys := make([]string, 0, len(stack.Hints()))
	for _, binding := range stack.Hints() {
		keys = append(keys, binding.Keys()...)
	}

	return keys
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func stackOf(overlays ...overlay.Overlay) overlay.Stack {
	stack := overlay.NewStack()
	for _, pushed := range overlays {
		stack, _, _ = stack.Pushed(pushed)
	}

	return stack
}
