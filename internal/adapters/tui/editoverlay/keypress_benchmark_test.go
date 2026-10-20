package editoverlay_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const typicalGoBlock = `func handle%d(w http.ResponseWriter, r *http.Request) {
	if err := process(r, %d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	fmt.Fprintf(w, "handled %d\n")
}

`

func BenchmarkSession_keypress(b *testing.B) {
	toStart := []tea.KeyPressMsg{{Code: tea.KeyHome, Mod: tea.ModCtrl}}
	tests := []struct {
		name    string
		moves   []tea.KeyPressMsg
		pressed tea.KeyPressMsg
	}{
		{name: "typing at the end of 256 KiB", moves: nil, pressed: keypress.Letter('x')},
		{name: "typing at the start of 256 KiB", moves: toStart, pressed: keypress.Letter('x')},
		{name: "breaking a line at the start of 256 KiB", moves: toStart, pressed: keypress.Special(tea.KeyEnter)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			stack := editingContentOf(b, typicalGoOfSize(maxContentBytes))
			stack = pressedOn(stack, tt.moves...)

			for b.Loop() {
				stack = pressedOn(stack, tt.pressed)
				_ = stack.Render()
			}
		})
	}
}

func editingContentOf(b *testing.B, content string) outcome.Stack {
	b.Helper()

	opened, _ := editoverlay.Editing(
		testsettings.Default(b).Keys,
		look.NewStyles(look.SchemeDark),
		editoverlay.Options{},
		editoverlay.BrowsedSnippet{Snippet: storedSnippet(b, content), Selection: browsed()},
	)
	stack, _, _ := outcome.NewStack().Update(tea.WindowSizeMsg{Width: screenSize().Width, Height: screenSize().Height})
	stack, _, _ = stack.Pushed(opened)

	return pressedOn(stack, enterContent()...)
}

func pressedOn(stack outcome.Stack, keys ...tea.KeyPressMsg) outcome.Stack {
	for _, pressed := range keys {
		stack, _, _ = stack.Update(pressed)
	}

	return stack
}

func typicalGoOfSize(size int) string {
	var source strings.Builder

	for index := 0; source.Len() < size; index++ {
		fmt.Fprintf(&source, typicalGoBlock, index, index, index)
	}

	return source.String()[:size]
}
