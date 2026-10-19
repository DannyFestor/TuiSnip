package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const overlongNotice = "Longer than 10,000 lines: read-only here"

type readOnlyContent struct {
	notice      string
	content     string
	language    value.Language
	highlighted string
}

func editableContent() readOnlyContent {
	return readOnlyContent{notice: "", content: "", language: value.Language{}, highlighted: ""}
}

func readOnlyIfOverlong(content string, language value.Language, codeStyle string) readOnlyContent {
	if !overflowsEmptyContent(content) {
		return editableContent()
	}

	held := readOnlyContent{notice: overlongNotice, content: content, language: language, highlighted: ""}

	return held.highlightedIn(codeStyle)
}

func (r readOnlyContent) held() bool {
	return r.notice != ""
}

func (r readOnlyContent) inLanguage(language value.Language, codeStyle string) readOnlyContent {
	relabelled := r
	relabelled.language = language

	return relabelled.highlightedIn(codeStyle)
}

func (r readOnlyContent) highlightedIn(codeStyle string) readOnlyContent {
	if !r.held() {
		return r
	}

	r.highlighted = look.Highlight(r.content, r.language.String(), codeStyle)

	return r
}

func (r readOnlyContent) text(externalEditorKey string) string {
	if externalEditorKey == "" {
		return r.notice
	}

	return r.notice + ", edit with " + externalEditorKey + " ($EDITOR)"
}
