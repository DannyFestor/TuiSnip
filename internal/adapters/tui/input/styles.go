package input

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

type areaRoles struct {
	text             lipgloss.Style
	cursorLine       lipgloss.Style
	cursorLineNumber lipgloss.Style
}

func LineStyles(styles look.Styles) textinput.Styles {
	return textinput.Styles{
		Focused: lineState(styles, styles.Plain),
		Blurred: lineState(styles, styles.Dim),
		Cursor: textinput.CursorStyle{
			Color:      styles.Accent.GetForeground(),
			Shape:      tea.CursorBlock,
			Blink:      false,
			BlinkSpeed: 0,
		},
	}
}

func AreaStyles(styles look.Styles) textarea.Styles {
	return textarea.Styles{
		Focused: areaState(styles, areaRoles{
			text:             styles.Plain,
			cursorLine:       styles.CurrentLine,
			cursorLineNumber: styles.Accent,
		}),
		Blurred: areaState(styles, areaRoles{text: styles.Dim, cursorLine: styles.Plain, cursorLineNumber: styles.Dim}),
		Cursor: textarea.CursorStyle{
			Color:      styles.Accent.GetForeground(),
			Shape:      tea.CursorBlock,
			Blink:      false,
			BlinkSpeed: 0,
		},
	}
}

func RestyledLine(line textinput.Model, styles look.Styles) textinput.Model {
	line.SetStyles(LineStyles(styles))

	return line
}

func RestyledArea(area textarea.Model, styles look.Styles) textarea.Model {
	area.SetStyles(AreaStyles(styles))

	return area
}

func lineState(styles look.Styles, text lipgloss.Style) textinput.StyleState {
	return textinput.StyleState{Text: text, Placeholder: styles.Dim, Suggestion: styles.Dim, Prompt: styles.Plain}
}

func areaState(styles look.Styles, roles areaRoles) textarea.StyleState {
	return textarea.StyleState{
		Base:             styles.Plain,
		Text:             roles.text,
		LineNumber:       styles.Dim,
		CursorLineNumber: roles.cursorLineNumber,
		CursorLine:       roles.cursorLine,
		EndOfBuffer:      styles.Dim,
		Placeholder:      styles.Dim,
		Prompt:           styles.Plain,
		Selection:        styles.Selection,
	}
}
