package keypress

import tea "charm.land/bubbletea/v2"

func Letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r), Mod: 0, ShiftedCode: 0, BaseCode: 0, IsRepeat: false}
}

func Special(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: "", Mod: 0, ShiftedCode: 0, BaseCode: 0, IsRepeat: false}
}

func Ctrl(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: "", Mod: tea.ModCtrl, ShiftedCode: 0, BaseCode: 0, IsRepeat: false}
}

func Typed(text string) []tea.KeyPressMsg {
	pressed := make([]tea.KeyPressMsg, 0, len(text))
	for _, r := range text {
		pressed = append(pressed, Letter(r))
	}

	return pressed
}
