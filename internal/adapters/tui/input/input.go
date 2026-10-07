package input

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func NewLine(prompt string, styles look.Styles) textinput.Model {
	line := textinput.New()
	line.Prompt = prompt
	line.KeyMap.Paste = withoutClipboardAccess(line.KeyMap.Paste)

	return RestyledLine(line, styles)
}

func NewContentArea(styles look.Styles) textarea.Model {
	area := textarea.New()
	area.Prompt = ""
	area.MaxHeight = 0
	area.KeyMap.Paste = withoutClipboardAccess(area.KeyMap.Paste)
	area.KeyMap.CopySelection = withoutClipboardAccess(area.KeyMap.CopySelection)

	return RestyledArea(area, styles)
}

func withoutClipboardAccess(binding key.Binding) key.Binding {
	binding.SetEnabled(false)

	return binding
}
