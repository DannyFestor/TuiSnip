package editoverlay

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"

func CaptureRefusal(keys binding.Keys, captured string) (string, bool) {
	if !overflowsEmptyContent(captured) {
		return "", false
	}

	return refusalText(captureOverflowsContent, formKeysOf(keys).externalEditorKey()), true
}
