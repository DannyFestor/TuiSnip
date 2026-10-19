package textarea

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Focused StyleState
	Blurred StyleState
	Cursor  CursorStyle
}

func DefaultStyles(isDark bool) Styles {
	lightDark := lipgloss.LightDark(isDark)

	var styles Styles

	styles.Focused = StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lightDark(lipgloss.Color("255"), lipgloss.Color("0"))),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("240"), lipgloss.Color("240"))),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("254"), lipgloss.Color("0"))),
		LineNumber:       lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("249"), lipgloss.Color("7"))),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Selection:        lipgloss.NewStyle().Background(lightDark(lipgloss.Color("253"), lipgloss.Color("8"))),
		Text:             lipgloss.NewStyle(),
	}
	styles.Blurred = StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("245"), lipgloss.Color("7"))),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("249"), lipgloss.Color("7"))),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("254"), lipgloss.Color("0"))),
		LineNumber:       lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("249"), lipgloss.Color("7"))),
		Placeholder:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Selection:        lipgloss.NewStyle().Background(lightDark(lipgloss.Color("253"), lipgloss.Color("8"))),
		Text:             lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("245"), lipgloss.Color("7"))),
	}
	styles.Cursor = CursorStyle{
		Color:      lipgloss.Color("7"),
		Shape:      tea.CursorBlock,
		Blink:      true,
		BlinkSpeed: 0,
	}

	return styles
}

func DefaultLightStyles() Styles {
	return DefaultStyles(false)
}

func DefaultDarkStyles() Styles {
	return DefaultStyles(true)
}
