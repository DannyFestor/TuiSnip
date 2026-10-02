package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
)

func newLineInput(prompt string) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt
	input.KeyMap.Paste = withoutClipboardAccess(input.KeyMap.Paste)

	styles := input.Styles()
	styles.Cursor.Blink = false
	input.SetStyles(styles)

	return input
}

func newContentArea() textarea.Model {
	area := textarea.New()
	area.Prompt = ""
	area.MaxHeight = 0
	area.KeyMap.Paste = withoutClipboardAccess(area.KeyMap.Paste)
	area.KeyMap.CopySelection = withoutClipboardAccess(area.KeyMap.CopySelection)

	styles := area.Styles()
	styles.Cursor.Blink = false
	area.SetStyles(styles)

	return area
}

func withoutClipboardAccess(binding key.Binding) key.Binding {
	binding.SetEnabled(false)

	return binding
}
