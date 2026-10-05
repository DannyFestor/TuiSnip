//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const capturedContent = "echo 'Grüße 🚀'\r\n\tprintf '%s'\n"

func TestCaptureReadsTheClipboardByteExact(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	home.PutOnClipboard(t, capturedContent)

	captured, err := app.Capture.Run(t.Context(), snippet.CaptureInput{})

	require.NoError(t, err)
	assert.Equal(t, capturedContent, captured.String())
}

func TestCaptureReadsWithThePlatformToolUnderOSC52(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, osc52Only)
	app := home.Start(t, testapp.RecordingTool)
	home.PutOnClipboard(t, capturedContent)

	captured, err := app.Capture.Run(t.Context(), snippet.CaptureInput{})

	require.NoError(t, err)
	assert.Equal(t, capturedContent, captured.String())
}

func TestCaptureRefusals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tool    testapp.ClipboardTool
		wantErr error
	}{
		{name: "an empty clipboard", tool: testapp.RecordingTool, wantErr: domain.ErrClipboardEmpty},
		{name: "no clipboard tool", tool: testapp.NoTool, wantErr: domain.ErrNoClipboardTool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, app := testapp.Start(t, tt.tool)

			_, err := app.Capture.Run(t.Context(), snippet.CaptureInput{})

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCreateFilesSnippetInFolderCarryingTag(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := seededFolder(t, app, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "docker"})
	oneliner := seededTag(t, app, testkit.TagSpec{ID: ids.NewTagID(), Name: "oneliner"})

	created := create(t, app, snippet.CreateInput{
		Title: "prune", Description: "", Language: plainText, Content: "docker system prune\n",
		FolderID: docker.ID(), Tags: []domain.Tag{oneliner},
	})

	inFolder, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: docker.ID(), Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)

	withTag, err := app.SnippetsWithTag.Run(t.Context(), browse.SnippetsWithTagInput{
		TagID: oneliner.ID(), Order: domain.SortOrderTitle,
	})
	require.NoError(t, err)

	assert.Equal(t, []domain.SnippetID{created.ID()}, idsOf(inFolder))
	assert.Equal(t, []domain.SnippetID{created.ID()}, idsOf(withTag))
}

func idsOf(snippets []domain.Snippet) []domain.SnippetID {
	listed := make([]domain.SnippetID, 0, len(snippets))
	for index := range snippets {
		listed = append(listed, snippets[index].ID())
	}

	return listed
}
