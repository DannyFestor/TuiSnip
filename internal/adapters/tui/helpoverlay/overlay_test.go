package helpoverlay_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestOverlay_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pressed tea.KeyPressMsg
	}{
		{name: "? closes", pressed: keypress.Letter('?')},
		{name: "esc closes", pressed: keypress.Special(tea.KeyEscape)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := helping(t)

			screen.Press(tt.pressed)

			assert.False(t, screen.IsOpen())
			assert.Empty(t, screen.Outcomes())
		})
	}

	t.Run("closes with the remapped help and back keys", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeGlobal][binding.Help] = []string{"f1"}
		keys[binding.ScopeGlobal][binding.Back] = []string{"backspace"}

		for _, pressed := range []tea.KeyPressMsg{keypress.Special(tea.KeyF1), keypress.Special(tea.KeyBackspace)} {
			screen := helpingWith(t, keys, wide())

			screen.Press(pressed)

			assert.False(t, screen.IsOpen(), "%s", pressed)
		}
	})

	t.Run("ignores other keys and pastes", func(t *testing.T) {
		t.Parallel()

		screen := helping(t)

		screen.Press(keypress.Letter('q'))
		screen.Send(tea.PasteMsg{Content: "?"})

		assert.True(t, screen.IsOpen())
		assert.Empty(t, screen.Outcomes())
	})
}

func TestOverlay_View(t *testing.T) {
	t.Parallel()

	t.Run("lists each Binding with its label under a Help title", func(t *testing.T) {
		t.Parallel()

		screen := helping(t)

		assert.Contains(t, screen.Screen(), "Help")
		assert.Regexp(t, `q\s+quit`, screen.Screen())
		assert.Regexp(t, `N\s+new Folder`, screen.Screen())
	})

	t.Run("names every key of a Binding", func(t *testing.T) {
		t.Parallel()

		screen := helping(t)

		assert.Regexp(t, `j, down\s+down`, screen.Screen())
	})

	t.Run("fits a screen narrower than the list", func(t *testing.T) {
		t.Parallel()

		narrow := look.Size{Width: 30, Height: 12}
		screen := helpingWith(t, testsettings.Default(t).Keys, narrow)

		assert.LessOrEqual(t, screen.TopBorderWidth(), narrow.Width)
	})

	t.Run("dims the ellipsis that marks columns cut off", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		light := look.NewStyles(look.SchemeLight)
		screen := helpingWith(t, keys, look.Size{Width: 30, Height: 12})

		screen.Send(look.Restyled{Styles: light})

		dimEllipsisCutBeforeReset := strings.TrimSuffix(light.Dim.Render(look.Ellipsis), ansi.ResetStyle)
		assert.Contains(t, screen.StyledScreen(), dimEllipsisCutBeforeReset)
	})

	t.Run("draws in the new Styles at the width it has", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		narrow := look.Size{Width: 30, Height: 12}
		light := look.NewStyles(look.SchemeLight)
		screen := helpingWith(t, keys, narrow)

		screen.Send(look.Restyled{Styles: light})

		assert.Equal(t, helpingStyled(t, keys, light, narrow).StyledScreen(), screen.StyledScreen())
	})
}

func TestOverlay_ShortHelp(t *testing.T) {
	t.Parallel()

	t.Run("hints closing with the help and back keys", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "? close · esc close", helping(t).Hints())
	})

	t.Run("hints the remapped keys", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeGlobal][binding.Help] = []string{"f1", "?"}
		keys[binding.ScopeGlobal][binding.Back] = []string{"backspace"}

		assert.Equal(t, "f1 close · backspace close", helpingWith(t, keys, wide()).Hints())
	})
}
