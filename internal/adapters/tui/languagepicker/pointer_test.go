package languagepicker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestPicker_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("picks the clicked Language and closes", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "Go", "Bash", "YAML"), language(t, "Go"))

		screen.Click("YAML")

		assert.Equal(t, []outcome.Outcome{outcome.LanguagePicked{Language: language(t, "YAML")}}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("closes without picking on a click outside", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, languages(t, "Go", "Bash"), language(t, "Go"))

		screen.Send(pointer.Clicked{At: pointer.Point{X: 0, Y: 0}, Double: false})

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})
}
