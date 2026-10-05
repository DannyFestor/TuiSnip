//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestCreatedSnippetListedAtRoot(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	created := create(t, app, snippet.CreateInput{Title: "curl json", Description: "POST", Content: "curl -d @-\n"})

	listed := listAtRoot(t, app)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID(), listed[0].ID())
	assert.Equal(t, "curl json", listed[0].Title().String())
	assert.Equal(t, "POST", listed[0].Description().String())
	assert.Equal(t, "curl -d @-\n", listed[0].FirstFragment().Content().String())
	assert.True(t, listed[0].AtRoot())
}

func TestCreateWithBlankTitleSavesNothing(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	_, err := app.Create.Run(t.Context(), snippet.CreateInput{Title: "   ", Description: "", Content: "ls"})

	fieldErrors := domain.FieldErrors(err)
	require.Len(t, fieldErrors, 1)
	assert.Equal(t, domain.FieldTitle, fieldErrors[0].Field)
	assert.Empty(t, listAtRoot(t, app))
}

func TestSnippetsSurviveRestart(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	first, err := home.Open(t, testapp.RecordingTool)
	require.NoError(t, err)
	created := create(t, first, snippet.CreateInput{Title: "kept", Description: "", Content: "echo kept"})
	require.NoError(t, first.Close())

	second := home.Start(t, testapp.RecordingTool)

	listed := listAtRoot(t, second)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID(), listed[0].ID())
}

func TestSearchFindsCreatedSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
	}{
		{name: "by fuzzy title", query: "crmbr"},
		{name: "by title without diacritics", query: "creme brulee"},
		{name: "by Description", query: "torch"},
		{name: "by content substring", query: "sugar on top"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, app := testapp.Start(t, testapp.RecordingTool)
			wanted := create(t, app, snippet.CreateInput{
				Title:       "Crème brûlée",
				Description: "Use the torch",
				Content:     "sprinkle sugar on top\n",
			})
			create(t, app, snippet.CreateInput{Title: "Unrelated", Description: "", Content: "ls -la"})

			hits, err := app.Query.Run(t.Context(), search.QueryInput{Text: tt.query})

			require.NoError(t, err)
			require.Len(t, hits, 1)
			assert.Equal(t, wanted.ID(), hits[0].Snippet().ID())
		})
	}
}

func TestBlankSearchFindsNothing(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	create(t, app, snippet.CreateInput{Title: "anything", Description: "", Content: "ls"})

	hits, err := app.Query.Run(t.Context(), search.QueryInput{Text: "  "})

	require.NoError(t, err)
	assert.Empty(t, hits)
}

func listAtRoot(t *testing.T, app *bootstrap.App) []domain.Snippet {
	t.Helper()

	listed, err := app.SnippetsInFolder.Run(t.Context(), browse.SnippetsInFolderInput{
		FolderID: domain.FolderID{},
		Order:    domain.SortOrderTitle,
	})
	require.NoError(t, err)

	return listed
}
