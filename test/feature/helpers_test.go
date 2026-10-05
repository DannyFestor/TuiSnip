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

const plainText = "plaintext"

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

func seededTag(t *testing.T, app *bootstrap.App, spec testkit.TagSpec) domain.Tag {
	t.Helper()

	tag := testkit.Tag(t, spec)
	testapp.SeedTag(t, app, tag)

	return tag
}

func seededSnippet(
	t *testing.T,
	app *bootstrap.App,
	ids *testkit.SequentialIDs,
	folderID domain.FolderID,
) domain.Snippet {
	t.Helper()

	return seededTaggedSnippet(t, app, ids, testkit.SnippetSpec{FolderID: folderID})
}

func seededTaggedSnippet(
	t *testing.T,
	app *bootstrap.App,
	ids *testkit.SequentialIDs,
	spec testkit.SnippetSpec,
) domain.Snippet {
	t.Helper()

	spec.ID = ids.NewSnippetID()
	spec.Fragment.ID = ids.NewFragmentID()
	seeded := testkit.Snippet(t, spec)
	testapp.SeedSnippet(t, app, seeded)

	return seeded
}
