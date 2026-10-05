package look_test

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestTheme_Scheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		theme look.Theme
		want  look.Scheme
	}{
		{name: "starts auto dark", theme: look.ThemeAuto, want: look.SchemeDark},
		{name: "forces light", theme: look.ThemeLight, want: look.SchemeLight},
		{name: "forces dark", theme: look.ThemeDark, want: look.SchemeDark},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.theme.Scheme())
		})
	}
}

func TestTheme_SchemeOn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		theme      look.Theme
		background color.Color
		want       look.Scheme
	}{
		{name: "auto follows a dark background", theme: look.ThemeAuto, background: color.Black, want: look.SchemeDark},
		{
			name:       "auto follows a light background",
			theme:      look.ThemeAuto,
			background: color.White,
			want:       look.SchemeLight,
		},
		{
			name:       "light ignores a dark background",
			theme:      look.ThemeLight,
			background: color.Black,
			want:       look.SchemeLight,
		},
		{
			name:       "dark ignores a light background",
			theme:      look.ThemeDark,
			background: color.White,
			want:       look.SchemeDark,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.theme.SchemeOn(tea.BackgroundColorMsg{Color: tt.background}))
		})
	}
}
