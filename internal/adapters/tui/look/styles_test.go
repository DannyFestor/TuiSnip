package look_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestNewStyles(t *testing.T) {
	t.Parallel()

	t.Run("colours each scheme differently", func(t *testing.T) {
		t.Parallel()

		dark := look.NewStyles(look.SchemeDark)
		light := look.NewStyles(look.SchemeLight)

		assert.NotEqual(t, dark.Focused.Border.Render("│"), light.Focused.Border.Render("│"))
		assert.NotEqual(t, dark.Dim.Render("x"), light.Dim.Render("x"))
		assert.NotEqual(t, dark.Accent.Render("x"), light.Accent.Render("x"))
		assert.NotEqual(t, dark.CurrentLine.Render("x"), light.CurrentLine.Render("x"))
		assert.NotEqual(t, dark.Selection.Render("x"), light.Selection.Render("x"))
	})

	t.Run("sets the selection apart from the current line", func(t *testing.T) {
		t.Parallel()

		for _, scheme := range []look.Scheme{look.SchemeDark, look.SchemeLight} {
			styles := look.NewStyles(scheme)

			assert.NotEqual(t, styles.CurrentLine.GetBackground(), styles.Selection.GetBackground())
		}
	})

	tests := []struct {
		name   string
		scheme look.Scheme
		want   string
	}{
		{name: "highlights code dark in the dark scheme", scheme: look.SchemeDark, want: "github-dark"},
		{name: "highlights code light in the light scheme", scheme: look.SchemeLight, want: "github"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, look.NewStyles(tt.scheme).CodeStyle)
		})
	}
}
