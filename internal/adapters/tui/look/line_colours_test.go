package look_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/screencolour"
)

func TestLineColours_Highlight(t *testing.T) {
	t.Parallel()

	codeStyle := look.NewStyles(look.SchemeDark).CodeStyle

	t.Run("colours a token as the Snippet pane does", func(t *testing.T) {
		t.Parallel()

		colours := look.ColourLines("func main() {}", "Go", codeStyle)

		keyword := rangeAt(t, colours.Highlight(0, nil), 0)

		assert.Equal(t, 4, keyword.End)
		assert.Equal(t, inSnippetPane("func main() {}").Foreground(t, "func"), foreground(t, keyword.Style))
	})

	t.Run("colours every line of a block comment as a comment", func(t *testing.T) {
		t.Parallel()

		content := "/* first\nsecond */\nx := 1"
		colours := look.ColourLines(content, "Go", codeStyle)

		second := rangeAt(t, colours.Highlight(1, nil), 0)

		assert.Equal(t, len("second */"), second.End)
		assert.Equal(t, inSnippetPane(content).Foreground(t, "second */"), foreground(t, second.Style))
	})

	t.Run("counts a tab as one rune", func(t *testing.T) {
		t.Parallel()

		colours := look.ColourLines("x\n\treturn", "Go", codeStyle)

		keyword := rangeAt(t, colours.Highlight(1, nil), 1)

		assert.Equal(t, 1+len("return"), keyword.End)
	})

	t.Run("counts a wide rune as one rune", func(t *testing.T) {
		t.Parallel()

		colours := look.ColourLines(`s := "界"; return`, "Go", codeStyle)

		keyword := rangeAt(t, colours.Highlight(0, nil), len([]rune(`s := "界"; `)))

		assert.Equal(t, len([]rune(`s := "界"; return`)), keyword.End)
	})

	t.Run("colours nothing on a line past the end", func(t *testing.T) {
		t.Parallel()

		colours := look.ColourLines("func main() {}", "Go", codeStyle)

		assert.Empty(t, colours.Highlight(5, nil))
	})

	t.Run("colours nothing without Content", func(t *testing.T) {
		t.Parallel()

		var colours look.LineColours

		assert.Empty(t, colours.Highlight(0, nil))
	})
}

func rangeAt(t *testing.T, ranges []lipgloss.Range, start int) lipgloss.Range {
	t.Helper()

	for index := range ranges {
		if ranges[index].Start == start {
			return ranges[index]
		}
	}

	require.Failf(t, "no range", "no range starts at %d in %v", start, ranges)

	return lipgloss.Range{}
}

func foreground(t *testing.T, style lipgloss.Style) string {
	t.Helper()

	const token = "token"

	return screencolour.Foreground(t, style.Render(token), token)
}

func inSnippetPane(content string) screencolour.InSnippetPane {
	return screencolour.InSnippetPane{Content: content, Language: "Go", Scheme: look.SchemeDark}
}
