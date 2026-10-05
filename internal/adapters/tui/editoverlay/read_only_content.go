package editoverlay

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const readOnlyNotice = "Contains tabs: read-only here"

type readOnlyContent struct {
	held        bool
	content     string
	language    value.Language
	highlighted string
}

func editableContent() readOnlyContent {
	return readOnlyContent{held: false, content: "", language: value.Language{}, highlighted: ""}
}

func readOnlyIfTabbed(content string, language value.Language, codeStyle string) readOnlyContent {
	if !strings.Contains(content, tabCharacter) {
		return editableContent()
	}

	held := readOnlyContent{held: true, content: content, language: language, highlighted: ""}

	return held.highlightedIn(codeStyle)
}

func (r readOnlyContent) inLanguage(language value.Language, codeStyle string) readOnlyContent {
	relabelled := r
	relabelled.language = language

	return relabelled.highlightedIn(codeStyle)
}

func (r readOnlyContent) highlightedIn(codeStyle string) readOnlyContent {
	if !r.held {
		return r
	}

	r.highlighted = look.Highlight(r.content, r.language.String(), codeStyle)

	return r
}

func readOnlyText(externalEditorKey string) string {
	if externalEditorKey == "" {
		return readOnlyNotice
	}

	return readOnlyNotice + ", edit with " + externalEditorKey + " ($EDITOR)"
}
