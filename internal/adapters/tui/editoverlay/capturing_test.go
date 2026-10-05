package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
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

	t.Run("highlights captured content with tabs in the Destination's Language", func(t *testing.T) {
		t.Parallel()

		inGo := capturingIn(t, destinationIn(t, "Go"), tabbedContent)
		plain := capturing(t, tabbedContent)

		assert.Contains(t, inGo.Screen(), "Language    Go")
		assert.Equal(t, withoutLanguageRow(plain.Screen()), withoutLanguageRow(inGo.Screen()))
		assert.NotEqual(t, withoutLanguageRow(plain.StyledScreen()), withoutLanguageRow(inGo.StyledScreen()))
	})

	t.Run("re-highlights captured content with tabs in a picked Language", func(t *testing.T) {
		t.Parallel()

		picked := capturing(t, tabbedContent)
		unpicked := capturing(t, tabbedContent)

		picked.Press(pickLanguage())
		picked.Press(keypress.Typed("bash")...)
		picked.Press(keypress.Special(tea.KeyEnter))

		inBash := capturingIn(t, destinationIn(t, "Bash"), tabbedContent)
		assert.Equal(t, inBash.StyledScreen(), picked.StyledScreen())
		assert.NotEqual(t, withoutLanguageRow(unpicked.StyledScreen()), withoutLanguageRow(picked.StyledScreen()))
	})
}

func TestCaptureRefusal(t *testing.T) {
	t.Parallel()

	const overlongClipboard = "Clipboard is longer than 10,000 lines"

	unbound := testsettings.Default(t).Keys
	unbound[binding.ScopeContent][binding.OpenInEditor] = []string{}

	tests := []struct {
		name     string
		keys     binding.Keys
		captured string
		want     string
		refused  bool
	}{
		{name: "accepts 10,000 lines", keys: testsettings.Default(t).Keys, captured: linesOf(10_000)},
		{
			name:     "refuses 10,001 lines and names the external editor key",
			keys:     testsettings.Default(t).Keys,
			captured: linesOf(10_001),
			want:     overlongClipboard + "; use ctrl+e to edit in $EDITOR",
			refused:  true,
		},
		{
			name:     "counts a carriage return and line feed as two line breaks",
			keys:     testsettings.Default(t).Keys,
			captured: linesJoinedBy(5_001, "\r\n"),
			want:     overlongClipboard + "; use ctrl+e to edit in $EDITOR",
			refused:  true,
		},
		{
			name:     "names no key when the external editor is unbound",
			keys:     unbound,
			captured: linesOf(10_001),
			want:     overlongClipboard,
			refused:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, refused := editoverlay.CaptureRefusal(tt.keys, tt.captured)

			assert.Equal(t, tt.refused, refused)
			assert.Equal(t, tt.want, got)
		})
	}
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

func TestCapturing_saveInDestinationLanguage(t *testing.T) {
	t.Parallel()

	screen := capturingIn(t, destinationIn(t, "Go"), capturedContent)

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
