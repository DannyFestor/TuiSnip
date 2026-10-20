package editoverlay_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
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

	t.Run("shows captured content with tabs editable", func(t *testing.T) {
		t.Parallel()

		screen := capturing(t, tabbedContent)

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "    return")
		assert.NotContains(t, screen.Screen(), "read-only")
	})

	t.Run("highlights read-only captured content in the Destination's Language", func(t *testing.T) {
		t.Parallel()

		inGo := capturingIn(t, destinationIn(t, "Go"), overlongGo())
		plain := capturing(t, overlongGo())

		assert.Contains(t, inGo.Screen(), "Language       Go")
		assert.Equal(t, withoutLanguageRow(plain.Screen()), withoutLanguageRow(inGo.Screen()))
		assert.NotEqual(t, withoutLanguageRow(plain.StyledScreen()), withoutLanguageRow(inGo.StyledScreen()))
	})

	t.Run("re-highlights read-only captured content in a picked Language", func(t *testing.T) {
		t.Parallel()

		picked := capturing(t, overlongGo())
		unpicked := capturing(t, overlongGo())

		picked.Press(pickLanguage())
		picked.Press(keypress.Typed("bash")...)
		picked.Press(keypress.Special(tea.KeyEnter))

		inBash := capturingIn(t, destinationIn(t, "Bash"), overlongGo())
		assert.Equal(t, inBash.StyledScreen(), picked.StyledScreen())
		assert.NotEqual(t, withoutLanguageRow(unpicked.StyledScreen()), withoutLanguageRow(picked.StyledScreen()))
	})
}

func overlongGo() string {
	return strings.Repeat("if x {\n}\n", 5_001)
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
	}{
		{name: "saves the captured content under the typed title", captured: capturedContent},
		{name: "keeps captured content with tabs byte for byte", captured: tabbedContent},
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

	screen := capturingIn(t, destinationIn(t, "Go"), capturedContent)

	screen.Press(keypress.Typed("Prune")...)
	screen.Press(save())

	want := outcome.SaveRequested{Input: inputIn("Prune", "", "Go", capturedContent)}
	assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
}

func TestCapturing_folderGoneElsewhere(t *testing.T) {
	t.Parallel()

	t.Run("offers to save at the Root when the Folder captured into is gone", func(t *testing.T) {
		t.Parallel()

		screen := capturedIntoGoneFolder(t)

		assert.Contains(t, screen.Screen(), goneElsewhereTitle)
		assert.Contains(t, screen.Screen(), folderGoneQuestion)
		assert.Len(t, screen.Outcomes(), 1, "only the save request, no failure")
	})

	t.Run("asks to save the captured content at the Root in its Language on yes", func(t *testing.T) {
		t.Parallel()

		screen := capturedIntoGoneFolder(t)
		screen.Press(keypress.Letter('y'))

		atRoot := inputIn("Prune", "", "Go", capturedContent)
		atRoot.FolderID = domain.FolderID{}
		assert.Equal(t, outcome.Outcome(outcome.SaveRequested{Input: atRoot}), screen.Outcomes()[1])
	})

	t.Run("keeps the capture open on no", func(t *testing.T) {
		t.Parallel()

		screen := capturedIntoGoneFolder(t)
		screen.Press(keypress.Letter('n'))

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.NotContains(t, screen.Screen(), folderGoneQuestion)
		assert.Len(t, screen.Outcomes(), 1)
	})
}

func capturedIntoGoneFolder(t *testing.T) *overlaytest.Driver {
	t.Helper()

	screen := capturingIn(t, destinationIn(t, "Go"), capturedContent)
	screen.Press(keypress.Typed("Prune")...)
	screen.Press(save())
	screen.Send(editoverlay.SaveFinished{
		Snippet: domain.Snippet{},
		Err:     fmt.Errorf("snippet.Create: %w", domain.ErrFolderGone),
	})

	return screen
}

func TestCapturing_cancel(t *testing.T) {
	t.Parallel()

	screen := capturing(t, capturedContent)

	screen.Press(keypress.Special(tea.KeyEscape))

	assert.Contains(t, screen.Screen(), discardQuestion)
	assert.True(t, screen.IsOpen())
}
