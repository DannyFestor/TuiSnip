//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func create(t *testing.T, app *bootstrap.App, in snippet.CreateInput) domain.Snippet {
	t.Helper()

	created, err := app.Create.Run(t.Context(), in)
	require.NoError(t, err)

	return created
}

func seededFolder(t *testing.T, app *bootstrap.App, spec testkit.FolderSpec) domain.Folder {
	t.Helper()

	folder := testkit.Folder(t, spec)
	testapp.SeedFolder(t, app, folder)

	return folder
}

func seededSnippet(
	t *testing.T,
	app *bootstrap.App,
	ids *testkit.SequentialIDs,
	folderID domain.FolderID,
) domain.Snippet {
	t.Helper()

	filed := testkit.Snippet(t, testkit.SnippetSpec{
		ID:       ids.NewSnippetID(),
		FolderID: folderID,
		Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID()},
	})
	testapp.SeedSnippet(t, app, filed)

	return filed
}
