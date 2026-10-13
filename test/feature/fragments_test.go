//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestSnippetWithThreeFragmentsKeepsTheirOrderAndTitles(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	seeded := threeFragments(t)

	stored := home.InsertSnippet(t, app, seeded)

	assert.Equal(t, seeded.Fragments(), stored.Fragments())
}

func TestCopyCopiesTheChosenFragment(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	stored := home.InsertSnippet(t, app, threeFragments(t))

	copyFragment(t, app, stored, stored.Fragments()[1])

	assert.Equal(t, "on: push", home.Copied(t))
}

func TestEditingTheTitleOfASnippetWithSeveralFragmentsKeepsThemAll(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	stored := home.InsertSnippet(t, app, threeFragments(t))

	_, err := app.Update.Run(
		t.Context(),
		editOf(stored, "Deploy the service", stored.FirstFragment().Content().String()),
	)
	require.NoError(t, err)

	edited := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, "Deploy the service", edited.Title().String())
	assert.Equal(t, stored.Fragments(), edited.Fragments())
}

func TestDuplicateKeepsEveryFragment(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	original := home.InsertSnippet(t, app, threeFragments(t))

	duplicate, err := app.Duplicate.Run(t.Context(), snippet.DuplicateInput{SnippetID: original.ID()})
	require.NoError(t, err)

	copied := testapp.StoredSnippet(t, app, duplicate.ID()).Fragments()
	require.Len(t, copied, 3)

	for i, fragment := range copied {
		assert.NotEqual(t, original.Fragments()[i].ID(), fragment.ID())
		assert.Equal(t, original.Fragments()[i].Title(), fragment.Title())
		assert.Equal(t, original.Fragments()[i].Language(), fragment.Language())
		assert.Equal(t, original.Fragments()[i].Content(), fragment.Content())
	}
}

func threeFragments(t *testing.T) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:    ids.NewSnippetID(),
		Title: "Deploy",
		Fragments: []testkit.FragmentSpec{
			{ID: ids.NewFragmentID(), Title: "main.go", Language: "Go", Content: "package main\n"},
			{ID: ids.NewFragmentID(), Language: "YAML", Content: "on: push\n"},
			{ID: ids.NewFragmentID(), Title: "README", Content: "Read me\n"},
		},
	})
}
