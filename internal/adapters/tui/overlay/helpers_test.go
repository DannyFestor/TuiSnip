package overlay_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

type (
	tickMsg struct{}
	step    = overlay.Step[string]
	stack   = overlay.Stack[string]
)

func overlayMock(t *testing.T) *MockOverlay[string] {
	t.Helper()

	mocked := NewMockOverlay[string](t)
	mocked.EXPECT().Update(mock.AnythingOfType("look.Resized")).Return(stay(mocked)).Maybe()

	return mocked
}

func terminalSizedMock(t *testing.T) *MockOverlay[string] {
	t.Helper()

	mocked := NewMockOverlay[string](t)
	mocked.EXPECT().Update(wholeTerminal()).Return(stay(mocked)).Once()
	mocked.EXPECT().Update(look.Resized{Box: look.Size{Width: 0, Height: 0}}).Return(stay(mocked)).Maybe()

	return mocked
}

func parentMock(t *testing.T) *MockParent[string] {
	t.Helper()

	mocked := NewMockParent[string](t)
	mocked.EXPECT().Update(mock.AnythingOfType("look.Resized")).Return(stay(mocked)).Maybe()

	return mocked
}

func baseMock(t *testing.T) *MockBase[string] {
	t.Helper()

	mocked := NewMockBase[string](t)
	mocked.EXPECT().Update(mock.AnythingOfType("look.Resized")).Return(stay(mocked)).Maybe()

	return mocked
}

func stay(next overlay.Overlay[string]) step {
	return overlay.Stay(next)
}

func closing() step {
	return overlay.Close[string]()
}

func emptyStack() stack {
	return overlay.NewStack[string]()
}

func hinting(name string) []key.Binding {
	return []key.Binding{key.NewBinding(key.WithKeys(name), key.WithHelp(name, name))}
}

func hintKeys(of stack) []string {
	keys := make([]string, 0, len(of.ShortHelp()))
	for _, binding := range of.ShortHelp() {
		keys = append(keys, binding.Keys()...)
	}

	return keys
}

func terminal() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: 80, Height: 24}
}

func wholeTerminal() look.Resized {
	return look.Resized{Box: look.Size{Width: 80, Height: 24}}
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func stackOf(overlays ...overlay.Overlay[string]) stack {
	built := emptyStack()
	for _, pushed := range overlays {
		built, _, _ = built.Pushed(pushed)
	}

	return built
}
