//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestATagCarriedOnlyByTrashedSnippetsCountsNone(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	trashed := testapp.SeedSnippet(t, app, snippet.CreateInput{Tags: []domain.Tag{golang}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{FolderID: docker.ID(), Tags: []domain.Tag{golang}})
	require.NoError(t, app.Delete.Run(t.Context(), snippet.DeleteInput{SnippetID: trashed.ID()}))
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	preview, err := app.PreviewDeleteTag.Run(t.Context(), tag.PreviewDeleteInput{TagID: golang.ID()})

	require.NoError(t, err)
	assert.Zero(t, preview.SnippetCount)
}

func TestMovingASnippetIntoAFolderInTheTrashFails(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	docker := testapp.SeedFolder(t, app, folder.CreateInput{Name: "docker"})
	moving := testapp.SeedSnippet(t, app, snippet.CreateInput{})
	require.NoError(t, app.DeleteFolder.Run(t.Context(), folder.DeleteInput{FolderID: docker.ID()}))

	_, err := app.Move.Run(t.Context(), snippet.MoveInput{SnippetID: moving.ID(), FolderID: docker.ID()})

	require.ErrorIs(t, err, domain.ErrFolderGone)
	assert.True(t, testapp.StoredSnippet(t, app, moving.ID()).AtRoot())
}
