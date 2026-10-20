package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
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

	screen := capturingIn(t, destinationInGo(t), capturedContent)
	screen.Press(keypress.Typed("Prune")...)
	screen.Press(save())
	screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errFolderGone})

	return screen
}

func TestCapturing_cancel(t *testing.T) {
	t.Parallel()

	screen := capturing(t, capturedContent)

	screen.Press(keypress.Special(tea.KeyEscape))

	assert.Contains(t, screen.Screen(), discardQuestion)
	assert.True(t, screen.IsOpen())
}
