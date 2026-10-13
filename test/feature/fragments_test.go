//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

type storedFragment struct {
	title    string
	language string
	content  string
}

func TestSnippetWithThreeFragmentsKeepsTheirOrderAndTitles(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	stored := seedThreeFragments(t, app)

	assert.Equal(t, []storedFragment{
		{title: "main.go", language: "Go", content: "package main\n"},
		{title: "", language: "YAML", content: "on: push\n"},
		{title: "README", language: plainText, content: "Read me\n"},
	}, storedFragmentsOf(stored))
}

func TestCopyCopiesTheChosenFragment(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	stored := seedThreeFragments(t, app)

	copyFragment(t, app, stored, stored.Fragments()[1])

	assert.Equal(t, "on: push", home.Copied(t))
}

func TestEditingTheTitleOfASnippetWithSeveralFragmentsKeepsThemAll(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := seedThreeFragments(t, app)

	_, err := app.Update.Run(
		t.Context(),
		editOf(stored, "Deploy the service", stored.FirstFragment().Content().String()),
	)
	require.NoError(t, err)

	edited := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, "Deploy the service", edited.Title().String())
	assert.Equal(t, stored.Fragments(), edited.Fragments())
}

func TestEditingFragmentsPersistsTheirOrderTitlesAndContent(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := seedThreeFragments(t, app)
	edit := editOf(stored, "Deploy", stored.FirstFragment().Content().String())
	mainGo, readme := edit.Fragments[0], edit.Fragments[2]
	readme.Title = "NOTES"
	readme.Content = "Read me first\n"
	added := snippet.FragmentInput{
		ID:       domain.FragmentID{},
		Title:    "deploy.sh",
		Language: "Bash",
		Content:  "make deploy\n",
	}
	edit.Fragments = []snippet.FragmentInput{readme, added, mainGo}

	_, err := app.Update.Run(t.Context(), edit)
	require.NoError(t, err)

	edited := testapp.StoredSnippet(t, app, stored.ID())
	assert.Equal(t, []storedFragment{
		{title: "NOTES", language: plainText, content: "Read me first\n"},
		{title: "deploy.sh", language: "Bash", content: "make deploy\n"},
		{title: "main.go", language: "Go", content: "package main\n"},
	}, storedFragmentsOf(edited))
	assert.Equal(t, stored.Fragments()[2].ID(), edited.Fragments()[0].ID())
	assert.Equal(t, stored.Fragments()[0], edited.Fragments()[2])
}

func TestSavingASnippetWithoutFragmentsIsRefused(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := seedThreeFragments(t, app)
	edit := editOf(stored, "Deploy", stored.FirstFragment().Content().String())
	edit.Fragments = nil

	_, err := app.Update.Run(t.Context(), edit)

	require.ErrorIs(t, err, domain.ErrNoFragment)
	assert.Equal(t, stored, testapp.StoredSnippet(t, app, stored.ID()))
}

func TestDuplicateKeepsEveryFragment(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	original := seedThreeFragments(t, app)

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

func seedThreeFragments(t *testing.T, app *bootstrap.App) domain.Snippet {
	t.Helper()

	return testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title: "Deploy",
		Fragments: []snippet.FragmentInput{
			{ID: domain.FragmentID{}, Title: "main.go", Language: "Go", Content: "package main\n"},
			{ID: domain.FragmentID{}, Title: "", Language: "YAML", Content: "on: push\n"},
			{ID: domain.FragmentID{}, Title: "README", Language: plainText, Content: "Read me\n"},
		},
	})
}

func storedFragmentsOf(stored domain.Snippet) []storedFragment {
	fragments := make([]storedFragment, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		fragments = append(fragments, storedFragment{
			title:    fragment.Title().String(),
			language: fragment.Language().String(),
			content:  fragment.Content().String(),
		})
	}

	return fragments
}
