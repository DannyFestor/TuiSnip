package move

import "charm.land/bubbles/v2/viewport"

func Scrolled(view viewport.Model, direction Direction) viewport.Model {
	switch direction {
	case Down:
		view.ScrollDown(1)
	case Up:
		view.ScrollUp(1)
	case Top:
		view.GotoTop()
	case Bottom:
		view.GotoBottom()
	case PageDown:
		view.PageDown()
	case PageUp:
		view.PageUp()
	case None:
	}

	return view
}

func Wheeled(view viewport.Model, lines int) viewport.Model {
	if lines < 0 {
		view.ScrollUp(-lines)
	} else {
		view.ScrollDown(lines)
	}

	return view
}
