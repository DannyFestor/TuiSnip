package look

import tea "charm.land/bubbletea/v2"

const Percent = 100

type Size struct {
	Width  int
	Height int
}

func SizeOf(screen tea.WindowSizeMsg) Size {
	return Size{Width: screen.Width, Height: screen.Height}
}

func (s Size) Inner() Size {
	return Size{Width: max(0, s.Width-BorderWidth), Height: max(0, s.Height-BorderWidth)}
}

func (s Size) Share(sharePercent int) Size {
	return Size{Width: s.Width * sharePercent / Percent, Height: s.Height * sharePercent / Percent}
}
