package testapp

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const defaultSnippetTitle = "Snippet"

func SeedFolder(t *testing.T, app *bootstrap.App, in folder.CreateInput) domain.Folder {
	t.Helper()

	created, err := app.CreateFolder.Run(t.Context(), in)
	require.NoError(t, err)

	return StoredFolder(t, app, created.ID())
}

func SeedTag(t *testing.T, app *bootstrap.App, in tag.CreateInput) domain.Tag {
	t.Helper()

	created, err := app.CreateTag.Run(t.Context(), in)
	require.NoError(t, err)

	return StoredTag(t, app, created.ID())
}

// SeedSnippet defaults a blank Title and Language, so a test states only the fields it is about.
func SeedSnippet(t *testing.T, app *bootstrap.App, in snippet.CreateInput) domain.Snippet {
	t.Helper()

	created, err := app.Create.Run(t.Context(), withSnippetDefaults(in))
	require.NoError(t, err)

	return StoredSnippet(t, app, created.ID())
}

func withSnippetDefaults(requested snippet.CreateInput) snippet.CreateInput {
	requested.Title = cmp.Or(requested.Title, defaultSnippetTitle)
	requested.Fragments = slices.Clone(requested.Fragments)

	if len(requested.Fragments) == 0 {
		requested.Fragments = []snippet.FragmentInput{{ID: domain.FragmentID{}, Title: "", Language: "", Content: ""}}
	}

	for index := range requested.Fragments {
		fragment := &requested.Fragments[index]
		fragment.Language = cmp.Or(fragment.Language, value.PlainText().String())
	}

	return requested
}
