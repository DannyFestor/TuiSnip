package testapp

import (
	_ "embed"
	"fmt"
	"testing"
)

type Editor string

const (
	ReplacingEditor Editor = "replacing"
	TabbingEditor   Editor = "tabbing"
	FailingEditor   Editor = "failing"
)

var (
	//go:embed testdata/replacing-editor.sh
	replacingEditorScript []byte
	//go:embed testdata/tabbing-editor.sh
	tabbingEditorScript []byte
	//go:embed testdata/failing-editor.sh
	failingEditorScript []byte
)

func (h *Home) UseEditor(t *testing.T, editor Editor) {
	t.Helper()

	path := h.installedScript(t, string(editor)+"-editor", editor.script())
	h.WriteConfig(t, fmt.Sprintf("editor = %q\n", path))
}

func (e Editor) script() []byte {
	switch e {
	case ReplacingEditor:
		return replacingEditorScript
	case TabbingEditor:
		return tabbingEditorScript
	case FailingEditor:
		return failingEditorScript
	}

	return nil
}
