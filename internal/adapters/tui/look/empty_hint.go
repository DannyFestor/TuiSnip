package look

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

const noSnippetsText = "No Snippets here."

func EmptyHint(styles Styles, hints []key.Binding) string {
	lines := []string{styles.Dim.Render(noSnippetsText), ""}

	for _, hint := range hints {
		if hint.Enabled() {
			lines = append(lines, hint.Help().Key+"  "+hint.Help().Desc)
		}
	}

	return strings.Join(lines, "\n")
}
