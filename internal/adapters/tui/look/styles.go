package look

import "charm.land/lipgloss/v2"

const (
	accentColor       = "#7D56F4"
	mutedColor        = "#666666"
	cursorTextColor   = "#FFFFFF"
	cursorUnfocusedBg = "#3A3A3A"
	invalidColor      = "#FF5F5F"
)

type Styles struct {
	Focused   FrameStyle
	Unfocused FrameStyle
	Dim       lipgloss.Style
	Bold      lipgloss.Style
	Plain     lipgloss.Style
	Invalid   lipgloss.Style
}

func NewStyles() Styles {
	accent := lipgloss.Color(accentColor)
	muted := lipgloss.Color(mutedColor)

	return Styles{
		Focused: FrameStyle{
			Border: lipgloss.NewStyle().Foreground(accent),
			Title:  lipgloss.NewStyle().Foreground(accent).Bold(true),
			Cursor: lipgloss.NewStyle().Foreground(lipgloss.Color(cursorTextColor)).Background(accent),
		},
		Unfocused: FrameStyle{
			Border: lipgloss.NewStyle().Foreground(muted),
			Title:  lipgloss.NewStyle().Foreground(muted),
			Cursor: lipgloss.NewStyle().Background(lipgloss.Color(cursorUnfocusedBg)),
		},
		Dim:     lipgloss.NewStyle().Foreground(muted),
		Bold:    lipgloss.NewStyle().Bold(true),
		Plain:   lipgloss.NewStyle(),
		Invalid: lipgloss.NewStyle().Foreground(lipgloss.Color(invalidColor)).Bold(true),
	}
}
