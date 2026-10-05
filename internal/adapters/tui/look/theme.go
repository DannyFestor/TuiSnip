package look

import tea "charm.land/bubbletea/v2"

type Theme int

const (
	ThemeAuto Theme = iota
	ThemeLight
	ThemeDark
)

func (t Theme) Scheme() Scheme {
	if t == ThemeLight {
		return SchemeLight
	}

	return SchemeDark
}

func (t Theme) SchemeOn(background tea.BackgroundColorMsg) Scheme {
	if t != ThemeAuto {
		return t.Scheme()
	}

	if background.IsDark() {
		return SchemeDark
	}

	return SchemeLight
}
