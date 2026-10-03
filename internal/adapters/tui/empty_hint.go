package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

const noSnippetsText = "No Snippets here."

func emptyHint(styles look.Styles, hints []key.Binding) string {
	lines := []string{styles.Dim.Render(noSnippetsText), ""}

	for _, hint := range hints {
		if hint.Enabled() {
			lines = append(lines, hint.Help().Key+"  "+hint.Help().Desc)
		}
	}

	return strings.Join(lines, "\n")
}
