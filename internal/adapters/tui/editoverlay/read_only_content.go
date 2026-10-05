package editoverlay

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const readOnlyNotice = "Contains tabs: read-only here"

type readOnlyContent struct {
	held        bool
	highlighted string
}

func readOnlyIfTabbed(fragment domain.Fragment, codeStyle string) readOnlyContent {
	content := fragment.Content().String()
	if !strings.Contains(content, tabCharacter) {
		return readOnlyContent{held: false, highlighted: ""}
	}

	return readOnlyContent{held: true, highlighted: look.Highlight(content, fragment.Language().String(), codeStyle)}
}

func readOnlyText(externalEditorKey string) string {
	if externalEditorKey == "" {
		return readOnlyNotice
	}

	return readOnlyNotice + ", edit with " + externalEditorKey + " ($EDITOR)"
}
