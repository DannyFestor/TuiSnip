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

	t.Run("shows captured content with tabs read-only", func(t *testing.T) {
		t.Parallel()

		screen := capturing(t, tabbedContent)

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), readOnlyNotice)
	})
}

func TestCapturing_save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		captured string
		sent     []tea.Msg
	}{
		{name: "saves the captured content under the typed title", captured: capturedContent},
		{
			name:     "keeps captured content with tabs byte for byte",
			captured: tabbedContent,
			sent: []tea.Msg{
				keypress.Special(tea.KeyDown),
				keypress.Special(tea.KeyDown),
				keypress.Special(tea.KeyEnter),
				keypress.Special(tea.KeyTab),
				tea.PasteMsg{Content: "pasted"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := capturing(t, tt.captured)

			screen.Press(keypress.Typed("Prune")...)

			for _, msg := range tt.sent {
				screen.Send(msg)
			}

			screen.Press(save())

			want := outcome.SaveRequested{Input: input("Prune", "", tt.captured)}
			assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		})
	}
}

func TestCapturing_cancel(t *testing.T) {
	t.Parallel()

	screen := capturing(t, capturedContent)

	screen.Press(keypress.Special(tea.KeyEscape))

	assert.Contains(t, screen.Screen(), discardQuestion)
	assert.True(t, screen.IsOpen())
}
