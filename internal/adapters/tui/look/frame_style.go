package look

import "charm.land/lipgloss/v2"

type FrameStyle struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Cursor lipgloss.Style
}

func (f FrameStyle) CursorOn(line string) string {
	return f.Cursor.Render(line)
}
