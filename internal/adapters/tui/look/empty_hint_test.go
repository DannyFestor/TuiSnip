package look_test

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestEmptyHint(t *testing.T) {
	t.Parallel()

	t.Run("lists each enabled hint under the notice", func(t *testing.T) {
		t.Parallel()

		disabled := key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help"))
		disabled.SetEnabled(false)
		hints := []key.Binding{key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new Snippet")), disabled}

		got := strings.Split(ansi.Strip(look.EmptyHint(look.NewStyles(), hints)), "\n")

		assert.Equal(t, []string{"No Snippets here.", "", "n  new Snippet"}, got)
	})

	t.Run("shows only the notice without hints", func(t *testing.T) {
		t.Parallel()

		got := strings.Split(ansi.Strip(look.EmptyHint(look.NewStyles(), nil)), "\n")

		assert.Equal(t, []string{"No Snippets here.", ""}, got)
	})
}
