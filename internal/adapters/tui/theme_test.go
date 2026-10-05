package tui_test

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestModel_theme(t *testing.T) {
	t.Parallel()

	t.Run("auto draws dark when the terminal reports no background", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, themed(t, look.ThemeDark).styledScreen(), themed(t, look.ThemeAuto).styledScreen())
	})

	tests := []struct {
		name       string
		theme      look.Theme
		background color.Color
		drawnAs    look.Theme
	}{
		{
			name:       "auto follows a light background",
			theme:      look.ThemeAuto,
			background: color.White,
			drawnAs:    look.ThemeLight,
		},
		{
			name:       "auto follows a dark background",
			theme:      look.ThemeAuto,
			background: color.Black,
			drawnAs:    look.ThemeDark,
		},
		{
			name:       "light ignores a dark background",
			theme:      look.ThemeLight,
			background: color.Black,
			drawnAs:    look.ThemeLight,
		},
		{
			name:       "dark ignores a light background",
			theme:      look.ThemeDark,
			background: color.White,
			drawnAs:    look.ThemeDark,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := themed(t, tt.theme)

			screen.send(tea.BackgroundColorMsg{Color: tt.background})

			assert.Equal(t, themed(t, tt.drawnAs).styledScreen(), screen.styledScreen())
		})
	}
}

func themed(t *testing.T, theme look.Theme) *driver {
	t.Helper()

	settings := testsettings.Default(t)
	settings.Theme = theme
	model := modelWithSettings(t, browsingActions(t, foldertree.New(t), listerOf(t, sampleSnippets(t)...)), settings)

	return start(t, model, wideWidth, wideHeight)
}
