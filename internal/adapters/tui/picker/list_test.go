package picker_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestList_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the filter line above every choice", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash", "YAML"))

		assert.Equal(t, []string{"filter:", "", "Go", "Bash", "YAML"}, visibleLines(list))
	})

	t.Run("shows each choice's meta at the right edge", func(t *testing.T) {
		t.Parallel()

		list := opened(t, []picker.Choice{{Text: "docker", Meta: "3"}})

		assert.Contains(t, visibleLines(list), "docker"+strings.Repeat(" ", boxWidth-len("docker")-1)+"3")
	})

	t.Run("shows each choice's mark before its text", func(t *testing.T) {
		t.Parallel()

		list := opened(t, []picker.Choice{
			{Mark: "✓ ", Text: "docker", Meta: "", Trailing: false},
			{Mark: "  ", Text: "go", Meta: "", Trailing: false},
		})

		assert.Equal(t, []string{"filter:", "", "✓ docker", "  go"}, visibleLines(list))
	})

	t.Run("shows the no-match text when the filter matches nothing", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Go", "Bash")), "zzz")

		assert.Equal(t, []string{"filter: zzz", "", noMatches}, visibleLines(list))
	})

	t.Run("scrolls a long filter to keep its end and the cursor in the box", func(t *testing.T) {
		t.Parallel()

		list := resized(opened(t, choicesOf("Go")), look.Size{Width: 20, Height: 4})
		list = typed(t, list, "abcdefghijklmnop")

		assert.Equal(t, "filter: fghijklmnop", visibleLines(list)[0])
	})

	t.Run("shows only the rows that fit and scrolls to the cursor", func(t *testing.T) {
		t.Parallel()

		list := resized(opened(t, choicesOf("A", "B", "C", "D")), look.Size{Width: boxWidth, Height: 4})
		list = pressed(t, list, down(), down(), down())

		assert.Equal(t, []string{"filter:", "", "C", "D"}, visibleLines(list))
	})
}

func TestList_Update(t *testing.T) {
	t.Parallel()

	t.Run("picks the first choice when nothing moved", func(t *testing.T) {
		t.Parallel()

		_, result := press(t, opened(t, choicesOf("Go", "Bash")), enter())

		assert.Equal(t, picked(0), result)
	})

	t.Run("moves down and up before picking", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, opened(t, choicesOf("Go", "Bash", "YAML")), down(), down(), up())

		_, result := press(t, list, enter())

		assert.Equal(t, picked(1), result)
	})

	t.Run("filters ignoring case, listing prefix matches first", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Algol", "Bash", "go", "Go Template")), "GO")

		assert.Equal(t, []string{"filter: GO", "", "go", "Go Template", "Algol"}, visibleLines(list))
	})

	t.Run("filters ignoring the space around the filter", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Bash", "Go")), " go ")

		assert.Equal(t, []string{"filter:  go", "", "Go"}, visibleLines(list))
	})

	t.Run("lists a trailing choice after the matches whatever the filter", func(t *testing.T) {
		t.Parallel()

		create := picker.Choice{Mark: "+ ", Text: "create", Meta: "", Trailing: true}
		list := typed(t, opened(t, append([]picker.Choice{create}, choicesOf("Go", "Bash")...)), "zzz")

		_, result := press(t, list, enter())

		assert.Equal(t, []string{"filter: zzz", "", "+ create"}, visibleLines(list))
		assert.Equal(t, picked(0), result)
	})

	t.Run("picks by the index in the full list after filtering", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Go", "Bash", "YAML")), "yam")

		_, result := press(t, list, enter())

		assert.Equal(t, picked(2), result)
	})

	t.Run("moves the cursor back to the top when the filter changes", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, opened(t, choicesOf("Bash", "Basic", "Go")), down())
		list = typed(t, list, "bas")

		_, result := press(t, list, enter())

		assert.Equal(t, picked(0), result)
	})

	t.Run("picks nothing when the filter matches nothing", func(t *testing.T) {
		t.Parallel()

		list := typed(t, opened(t, choicesOf("Go")), "zzz")

		_, result := press(t, list, enter())

		assert.Equal(t, picker.Filtering, result.Ending)
	})

	t.Run("reports cancel", func(t *testing.T) {
		t.Parallel()

		_, result := press(t, opened(t, choicesOf("Go")), keypress.Special(tea.KeyEscape))

		assert.Equal(t, picker.Cancelled, result.Ending)
	})

	t.Run("reports typing as filtering", func(t *testing.T) {
		t.Parallel()

		_, result := press(t, opened(t, choicesOf("Go")), keypress.Letter('g'))

		assert.Equal(t, picker.Filtering, result.Ending)
	})

	t.Run("filters on pasted text", func(t *testing.T) {
		t.Parallel()

		list, _, _ := opened(t, choicesOf("Go", "Bash")).Update(tea.PasteMsg{Content: "ba"})

		assert.Equal(t, []string{"filter: ba", "", "Bash"}, visibleLines(list))
	})

	t.Run("uses the configured keys", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopePicker][binding.Accept] = []string{"ctrl+o"}
		keys[binding.ScopePicker][binding.Down] = []string{"ctrl+d"}
		list := openedWith(t, keys, choicesOf("Go", "Bash"))

		list = pressed(t, list, keypress.Ctrl('d'))
		_, result := press(t, list, keypress.Ctrl('o'))

		assert.Equal(t, picked(1), result)
	})
}

func TestList_WithCursorOn(t *testing.T) {
	t.Parallel()

	t.Run("puts the cursor on the named choice", func(t *testing.T) {
		t.Parallel()

		list := opened(t, choicesOf("Go", "Bash", "YAML")).WithCursorOn("YAML")

		_, result := press(t, list, enter())

		assert.Equal(t, picked(2), result)
	})

	t.Run("puts the cursor back on the first choice", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, opened(t, choicesOf("Go", "Bash")), down()).WithCursorOn("Go")

		_, result := press(t, list, enter())

		assert.Equal(t, picked(0), result)
	})

	t.Run("leaves the cursor where it is for a choice not shown", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, opened(t, choicesOf("Go", "Bash")), down()).WithCursorOn("Rust")

		_, result := press(t, list, enter())

		assert.Equal(t, picked(1), result)
	})
}

func TestList_WithChoices(t *testing.T) {
	t.Parallel()

	list := typed(t, opened(t, choicesOf("Go")), "b").WithChoices(choicesOf("Go", "Bash", "BibTeX"))

	assert.Equal(t, []string{"filter: b", "", "Bash", "BibTeX"}, visibleLines(list))
}

func TestList_Highlighted(t *testing.T) {
	t.Parallel()

	t.Run("names the choice under the cursor", func(t *testing.T) {
		t.Parallel()

		highlighted, ok := pressed(t, opened(t, choicesOf("Go", "Bash")), down()).Highlighted()

		assert.True(t, ok)
		assert.Equal(t, picker.Choice{Mark: "", Text: "Bash", Meta: "", Trailing: false}, highlighted)
	})

	t.Run("names nothing when the filter matches nothing", func(t *testing.T) {
		t.Parallel()

		_, ok := typed(t, opened(t, choicesOf("Go")), "zzz").Highlighted()

		assert.False(t, ok)
	})
}

func TestList_Filter(t *testing.T) {
	t.Parallel()

	list := typed(t, opened(t, choicesOf("Go")), " te ")

	assert.Equal(t, " te ", list.Filter())
}

func TestList_ShortHelp(t *testing.T) {
	t.Parallel()

	list := opened(t, choicesOf("Go"))

	assert.Equal(t, "down move · enter pick · esc close", hintsOf(list))
}

func hintsOf(list picker.List) string {
	entries := make([]string, 0)

	for _, hint := range list.ShortHelp() {
		if hint.Enabled() {
			entries = append(entries, hint.Help().Key+" "+hint.Help().Desc)
		}
	}

	return strings.Join(entries, " · ")
}

func visibleLines(list picker.List) []string {
	lines := strings.Split(ansi.Strip(list.View(look.NewStyles(look.SchemeDark))), "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " ")
	}

	return lines
}
