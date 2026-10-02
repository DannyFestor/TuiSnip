package tui

import "charm.land/lipgloss/v2"

const (
	accentColor       = "#7D56F4"
	mutedColor        = "#666666"
	cursorTextColor   = "#FFFFFF"
	cursorUnfocusedBg = "#3A3A3A"
)

type paneLook struct {
	border lipgloss.Style
	title  lipgloss.Style
	cursor lipgloss.Style
}

type styleSet struct {
	focused   paneLook
	unfocused paneLook
	dim       lipgloss.Style
	bold      lipgloss.Style
}

func newStyleSet() styleSet {
	accent := lipgloss.Color(accentColor)
	muted := lipgloss.Color(mutedColor)

	return styleSet{
		focused: paneLook{
			border: lipgloss.NewStyle().Foreground(accent),
			title:  lipgloss.NewStyle().Foreground(accent).Bold(true),
			cursor: lipgloss.NewStyle().Foreground(lipgloss.Color(cursorTextColor)).Background(accent),
		},
		unfocused: paneLook{
			border: lipgloss.NewStyle().Foreground(muted),
			title:  lipgloss.NewStyle().Foreground(muted),
			cursor: lipgloss.NewStyle().Background(lipgloss.Color(cursorUnfocusedBg)),
		},
		dim:  lipgloss.NewStyle().Foreground(muted),
		bold: lipgloss.NewStyle().Bold(true),
	}
}

func (s styleSet) lookFor(p, focus pane) paneLook {
	if p == focus {
		return s.focused
	}

	return s.unfocused
}
