package editoverlay

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	tabbedNotice   = "Contains tabs: read-only here"
	overlongNotice = "Longer than 10,000 lines: read-only here"
)

type uneditableCheck func(content string) (notice string)

type readOnlyContent struct {
	notice      string
	content     string
	language    value.Language
	highlighted string
}

func editableContent() readOnlyContent {
	return readOnlyContent{notice: "", content: "", language: value.Language{}, highlighted: ""}
}

func readOnlyIfTabbed(content string, language value.Language, codeStyle string) readOnlyContent {
	return readOnlyIf(content, language, codeStyle, noticeIfTabbed)
}

func readOnlyIfUneditable(content string, language value.Language, codeStyle string) readOnlyContent {
	return readOnlyIf(content, language, codeStyle, noticeIfTabbed, noticeIfOverlong)
}

func readOnlyIf(
	content string, language value.Language, codeStyle string, checks ...uneditableCheck,
) readOnlyContent {
	for _, check := range checks {
		if notice := check(content); notice != "" {
			held := readOnlyContent{notice: notice, content: content, language: language, highlighted: ""}

			return held.highlightedIn(codeStyle)
		}
	}

	return editableContent()
}

func noticeIfTabbed(content string) string {
	if strings.Contains(content, tabCharacter) {
		return tabbedNotice
	}

	return ""
}

func noticeIfOverlong(content string) string {
	if overflowsEmptyContent(content) {
		return overlongNotice
	}

	return ""
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
