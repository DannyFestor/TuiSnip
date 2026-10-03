package confirm_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestConfirmation_Update(t *testing.T) {
	t.Parallel()

	t.Run("y reports the outcome it was opened with and closes", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		screen.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.DiscardConfirmed{}}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("n closes without an outcome", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		screen.Press(keypress.Letter('n'))

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("ignores other keys and pastes", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		screen.Press(keypress.Letter('z'))
		screen.Send(tea.PasteMsg{Content: "y"})

		assert.Empty(t, screen.Outcomes())
		assert.Contains(t, screen.Screen(), question)
	})
}

func TestConfirmation_View(t *testing.T) {
	t.Parallel()

	t.Run("asks with the answers, No as the default", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		assert.Contains(t, screen.Screen(), "Unsaved changes")
		assert.Contains(t, screen.Screen(), question+" y/N")
	})

	t.Run("fits the frame to a question wider than the title", func(t *testing.T) {
		t.Parallel()

		screen := asking(t)

		assert.Equal(t, len(question+" y/N")+2, screen.TopBorderWidth())
	})

	t.Run("keeps the frame wide enough for the title over a short question", func(t *testing.T) {
		t.Parallel()

		screen := askingWith(t, testsettings.Default(t).Keys, "Go?")

		assert.Equal(t, len(" Unsaved changes ")+2, screen.TopBorderWidth())
	})
}

func TestConfirmation_answers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yes  []string
		no   []string
		want string
	}{
		{name: "upper-cases a single-character No key", yes: []string{"y"}, no: []string{"x", "esc"}, want: "y/X"},
		{
			name: "writes a longer No key as configured",
			yes:  []string{"ctrl+y"},
			no:   []string{"shift+tab", "n"},
			want: "ctrl+y/shift+tab",
		},
		{name: "names only Yes when No is unbound", yes: []string{"y"}, no: []string{}, want: "y"},
		{name: "names only No when Yes is unbound", yes: []string{}, no: []string{"n"}, want: "N"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keys := testsettings.Default(t).Keys
			keys[binding.ScopeConfirm][binding.Yes] = tt.yes
			keys[binding.ScopeConfirm][binding.No] = tt.no

			screen := askingWith(t, keys, question)

			assert.Contains(t, screen.Screen(), "│"+question+" "+tt.want+"│")
		})
	}

	t.Run("asks without answers when yes and no are unbound", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeConfirm][binding.Yes] = []string{}
		keys[binding.ScopeConfirm][binding.No] = []string{}

		screen := askingWith(t, keys, question)

		assert.Contains(t, screen.Screen(), "│"+question+"│")
	})
}

func TestConfirmation_ShortHelp(t *testing.T) {
	t.Parallel()

	screen := asking(t)

	assert.Equal(t, "y yes · n no", screen.Hints())
}
