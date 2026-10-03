package look_test

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestHighlight(t *testing.T) {
	t.Parallel()

	t.Run("colours code in a known Language", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("func main() {}", "Go", look.DarkCodeStyle)

		assert.NotEqual(t, "func main() {}", got)
		assert.Equal(t, "func main() {}", ansi.Strip(got))
	})

	t.Run("keeps the text of an unknown Language", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("plain words", "NoSuchLanguage", look.DarkCodeStyle)

		assert.Equal(t, "plain words", ansi.Strip(got))
	})

	t.Run("shows tabs as four spaces and drops the final newline", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("if x {\n\treturn\n}\n", "Go", look.DarkCodeStyle)

		assert.Equal(t, "if x {\n    return\n}", ansi.Strip(got))
	})
}

func TestCodeStyleFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		background color.Color
		want       string
	}{
		{name: "picks the dark style on a dark background", background: color.Black, want: look.DarkCodeStyle},
		{name: "picks the light style on a light background", background: color.White, want: "github"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, look.CodeStyleFor(tea.BackgroundColorMsg{Color: tt.background}))
		})
	}
}
