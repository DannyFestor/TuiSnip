package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

const noSnippetsText = "No Snippets here."

func emptyHint(styles styleSet, hints []key.Binding) string {
	lines := []string{styles.dim.Render(noSnippetsText), ""}

	for _, hint := range hints {
		if hint.Enabled() {
			lines = append(lines, hint.Help().Key+"  "+hint.Help().Desc)
		}
	}

	return strings.Join(lines, "\n")
}
