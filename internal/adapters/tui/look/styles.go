package look

import "charm.land/lipgloss/v2"

type Styles struct {
	Focused   FrameStyle
	Unfocused FrameStyle
	Dim       lipgloss.Style
	Bold      lipgloss.Style
	Plain     lipgloss.Style
	Invalid   lipgloss.Style
	CodeStyle string
}

func NewStyles(scheme Scheme) Styles {
	colours := paletteOf(scheme)
	accent := lipgloss.Color(colours.accent)
	muted := lipgloss.Color(colours.muted)

	return Styles{
		Focused: FrameStyle{
			Border: lipgloss.NewStyle().Foreground(accent),
			Title:  lipgloss.NewStyle().Foreground(accent).Bold(true),
			Cursor: lipgloss.NewStyle().Foreground(lipgloss.Color(colours.cursorText)).Background(accent),
		},
		Unfocused: FrameStyle{
			Border: lipgloss.NewStyle().Foreground(muted),
			Title:  lipgloss.NewStyle().Foreground(muted),
			Cursor: lipgloss.NewStyle().Background(lipgloss.Color(colours.unfocusedCursor)),
		},
		Dim:       lipgloss.NewStyle().Foreground(muted),
		Bold:      lipgloss.NewStyle().Bold(true),
		Plain:     lipgloss.NewStyle(),
		Invalid:   lipgloss.NewStyle().Foreground(lipgloss.Color(colours.invalid)).Bold(true),
		CodeStyle: colours.codeStyle,
	}
}
