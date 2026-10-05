package languagepicker_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPicker_View(t *testing.T) {
	t.Parallel()

	t.Run("shows every Language ignoring case when config curates none", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, nil, language(t, "ABAP"))

		assert.Contains(t, screen.Screen(), pickerTitle)
		assert.Regexp(t, `(?s)ABAP.*\n.*ABNF.*\n.*ActionScript`, screen.Screen())
	})

	t.Run("scrolls to the current Language", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, nil, language(t, "YAML"))

		assert.Contains(t, screen.Screen(), "YAML")
		assert.NotContains(t, screen.Screen(), "ABAP")
	})

	t.Run("shows only the curated Languages in their order", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), value.PlainText())

		assert.Regexp(t, `(?s)YAML.*\n.*Go`, screen.Screen())
		assert.NotContains(t, screen.Screen(), "ABAP")
	})

	t.Run("reveals every Language and back with show_all_languages", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), value.PlainText())

		screen.Press(showAll())
		assert.Contains(t, screen.Screen(), "XML")

		screen.Press(showAll())
		assert.NotContains(t, screen.Screen(), "XML")
		assert.Contains(t, screen.Screen(), "YAML")
	})

	t.Run("reveals every Language with the configured key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopePicker][binding.ShowAllLanguages] = []string{"ctrl+o"}
		screen := pickingWith(t, keys, languages(t, "YAML"), value.PlainText())

		screen.Press(keypress.Ctrl('o'))

		assert.Contains(t, screen.Screen(), "XML")
	})

	t.Run("show_all_languages does nothing when config curates none", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, nil, language(t, "ABAP"))

		screen.Press(showAll())

		assert.Contains(t, screen.Screen(), "ABAP")
		assert.True(t, screen.IsOpen())
	})

	t.Run("keeps the filter when it reveals every Language", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), value.PlainText())
		screen.Press(keypress.Typed("bas")...)

		screen.Press(showAll())

		assert.Contains(t, screen.Screen(), "Bash")
		assert.NotContains(t, screen.Screen(), "YAML")
	})

	t.Run("hints move, pick and close", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, nil, value.PlainText())

		assert.Equal(t, "down move · enter pick · esc close", screen.Hints())
	})
}

func TestPicker_Update(t *testing.T) {
	t.Parallel()

	t.Run("picks the highlighted Language and closes", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), value.PlainText())

		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{pickedOutcome(t, "Go")}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("opens on the current Language", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go", "Bash"), language(t, "Bash"))

		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{pickedOutcome(t, "Bash")}, screen.Outcomes())
	})

	t.Run("picks from the filtered Languages", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, nil, value.PlainText())

		screen.Press(keypress.Typed("yam")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{pickedOutcome(t, "YAML")}, screen.Outcomes())
	})

	t.Run("keeps the highlighted Language when it reveals every Language", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), language(t, "Go"))

		screen.Press(showAll(), keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{pickedOutcome(t, "Go")}, screen.Outcomes())
	})

	t.Run("closes without picking", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML"), value.PlainText())

		screen.Press(keypress.Special(tea.KeyEscape))

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("stays open when nothing matches the filter", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML"), value.PlainText())

		screen.Press(keypress.Typed("zzz")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.Screen(), noMatches)
		assert.Empty(t, screen.Outcomes())
		assert.True(t, screen.IsOpen())
	})

	t.Run("filters on pasted text", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "YAML", "Go"), value.PlainText())

		screen.Send(tea.PasteMsg{Content: "go"})
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{pickedOutcome(t, "Go")}, screen.Outcomes())
	})
}
