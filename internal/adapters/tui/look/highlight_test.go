package look_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestHighlight(t *testing.T) {
	t.Parallel()

	codeStyle := look.NewStyles(look.SchemeDark).CodeStyle

	t.Run("colours code in a known Language", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("func main() {}", "Go", codeStyle)

		assert.NotEqual(t, "func main() {}", got)
		assert.Equal(t, "func main() {}", ansi.Strip(got))
	})

	t.Run("keeps the text of an unknown Language", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("plain words", "NoSuchLanguage", codeStyle)

		assert.Equal(t, "plain words", ansi.Strip(got))
	})

	t.Run("runs tabs to the next tab stop and drops the final newline", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("if x {\n\treturn\n}\n", "Go", codeStyle)

		assert.Equal(t, "if x {\n    return\n}", ansi.Strip(got))
	})

	t.Run("runs a tab after one cell to cell 4", func(t *testing.T) {
		t.Parallel()

		got := look.Highlight("a\tb", "NoSuchLanguage", codeStyle)

		assert.Equal(t, "a   b", ansi.Strip(got))
	})
}
