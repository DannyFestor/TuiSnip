package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const capturedContent = "docker system prune\n"

func TestCapturing_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the captured content as an unsaved change", func(t *testing.T) {
		t.Parallel()

		screen := capturing(t, capturedContent)

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "docker system prune")
	})

	t.Run("shows captured content with tabs editable", func(t *testing.T) {
		t.Parallel()

		screen := capturing(t, tabbedContent)

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "    return")
		assert.NotContains(t, screen.Screen(), "read-only")
	})
}

func TestCapturing_save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		captured string
	}{
		{name: "saves the captured content under the typed title", captured: capturedContent},
		{name: "keeps captured content with tabs byte for byte", captured: tabbedContent},
		{name: "keeps a capture of 20,000 short lines byte for byte", captured: twentyThousandShortLines()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := capturing(t, tt.captured)

			screen.Press(keypress.Typed("Prune")...)
			screen.Press(save())

			want := outcome.SaveRequested{Input: input("Prune", "", tt.captured)}
			assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		})
	}
}

func TestCapturing_saveInDestinationLanguage(t *testing.T) {
	t.Parallel()

	screen := capturingIn(t, destinationInGo(t), capturedContent)

	screen.Press(keypress.Typed("Prune")...)
	screen.Press(save())

	want := outcome.SaveRequested{Input: inputIn("Prune", "", "Go", capturedContent)}
	assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
}

func TestCapturing_cancel(t *testing.T) {
	t.Parallel()

	screen := capturing(t, capturedContent)

	screen.Press(keypress.Special(tea.KeyEscape))

	assert.Contains(t, screen.Screen(), discardQuestion)
	assert.True(t, screen.IsOpen())
}
