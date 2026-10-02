package tui

const (
	editOverlayPercent = 90
	searchPopupPercent = 80
)

func shareOf(screen size, sharePercent int) size {
	return size{width: screen.width * sharePercent / percent, height: screen.height * sharePercent / percent}
}
