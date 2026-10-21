package screencolour

import (
	"testing"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

type InSnippetPane struct {
	Content  string
	Language string
	Scheme   look.Scheme
}

func (p InSnippetPane) Foreground(tb testing.TB, text string) string {
	tb.Helper()

	coloured := Foreground(tb, look.Highlight(p.Content, p.Language, look.NewStyles(p.Scheme).CodeStyle), text)
	if coloured == "" {
		tb.Fatalf("the Snippet pane doesn't colour %q", text)
	}

	return coloured
}
