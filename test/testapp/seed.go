package testapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func SeedFolder(t *testing.T, app *bootstrap.App, folder domain.Folder) {
	t.Helper()

	require.NoError(t, app.FolderRepository.Insert(t.Context(), folder))
}

func SeedTag(t *testing.T, app *bootstrap.App, tag domain.Tag) {
	t.Helper()

	require.NoError(t, app.TagRepository.Insert(t.Context(), tag))
}

func SeedSnippet(t *testing.T, app *bootstrap.App, snippet domain.Snippet) {
	t.Helper()

	require.NoError(t, app.SnippetRepository.Insert(t.Context(), snippet))
}
