package editoverlay

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const readOnlyNotice = "Contains tabs: read-only here"

type readOnlyContent struct {
	held        bool
	fragment    domain.Fragment
	highlighted string
}

func editableContent() readOnlyContent {
	return readOnlyContent{held: false, fragment: domain.Fragment{}, highlighted: ""}
}

func readOnlyIfTabbed(fragment domain.Fragment, codeStyle string) readOnlyContent {
	if !strings.Contains(fragment.Content().String(), tabCharacter) {
		return editableContent()
	}

	held := readOnlyContent{held: true, fragment: fragment, highlighted: ""}

	return held.highlightedIn(codeStyle)
}

func (r readOnlyContent) highlightedIn(codeStyle string) readOnlyContent {
	if !r.held {
		return r
	}

	r.highlighted = look.Highlight(r.fragment.Content().String(), r.fragment.Language().String(), codeStyle)

	return r
}

func readOnlyText(externalEditorKey string) string {
	if externalEditorKey == "" {
		return readOnlyNotice
	}

	return readOnlyNotice + ", edit with " + externalEditorKey + " ($EDITOR)"
}
