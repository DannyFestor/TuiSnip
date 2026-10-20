package editoverlay_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencolour"
)

func TestEditing_highlighting(t *testing.T) {
	t.Parallel()

	t.Run("shows the Snippet pane's colours, block comments included", func(t *testing.T) {
		t.Parallel()

		const content = "/* first\nsecond */\nfunc main() {}"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()

		for _, text := range []string{"second */", "func"} {
			assert.Equal(t,
				snippetPaneForeground(t, look.SchemeDark, content, text),
				screencolour.Foreground(t, screen.StyledScreen(), text),
				text,
			)
		}
	})

	t.Run("keeps the old colours until typing pauses, then updates them", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}\n"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("// note")...)

		assert.Equal(t,
			snippetPaneForeground(t, look.SchemeDark, content, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "// note"))

		screen.RunCommands()

		assert.Equal(t,
			snippetPaneForeground(t, look.SchemeDark, content+"// note", "// note"),
			screencolour.Foreground(t, screen.StyledScreen(), "// note"),
		)
	})

	t.Run("shows plain text without colours", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippetIn(t, "plaintext", "func main() {}"))
		screen.RunCommands()

		assert.Empty(t, screencolour.Foreground(t, screen.StyledScreen(), "func"))
	})

	t.Run("colours the Content again in a picked Language", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}"

		screen := editingStored(t, storedSnippetIn(t, "plaintext", content))
		screen.RunCommands()
		screen.Offer(outcome.LanguagePicked{Language: language(t, "Go")})
		screen.RunCommands()

		assert.Equal(t,
			snippetPaneForeground(t, look.SchemeDark, content, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
	})

	t.Run("colours the Content in the code style of a new scheme", func(t *testing.T) {
		t.Parallel()

		const content = "func main() {}"

		screen := editingStored(t, storedSnippet(t, content))
		screen.RunCommands()
		screen.Send(look.Restyled{Styles: look.NewStyles(look.SchemeLight)})
		screen.RunCommands()

		assert.Equal(t,
			snippetPaneForeground(t, look.SchemeLight, content, "func"),
			screencolour.Foreground(t, screen.StyledScreen(), "func"),
		)
	})
}

func snippetPaneForeground(t *testing.T, scheme look.Scheme, content, text string) string {
	t.Helper()

	highlighted := look.Highlight(content, "Go", look.NewStyles(scheme).CodeStyle)
	coloured := screencolour.Foreground(t, highlighted, text)
	require.NotEmpty(t, coloured, "the Snippet pane doesn't colour %q", text)

	return coloured
}
