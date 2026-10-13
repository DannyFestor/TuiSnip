package snippet_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewUpdate(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewUpdate(nil, nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "snippet.NewUpdate: repo")
		require.ErrorContains(t, err, "ids")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewUpdate(NewMockUpdateRepository(t), testkit.NewSequentialIDs(), fixedClock())

		assert.NoError(t, err)
	})
}

func TestUpdate_Run(t *testing.T) {
	t.Parallel()

	t.Run("saves the edited fields guarded by the loaded time", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "curl")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		var saved domain.Snippet

		repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("domain.Snippet"), loadedAt()).
			Run(func(_ context.Context, s domain.Snippet, _ time.Time) { saved = s }).
			Return(nil)

		updated, err := newUpdate(t, repo).Run(t.Context(), updateInput(stored, " curl json ", "POST", "curl -d\n"))

		require.NoError(t, err)
		assert.Equal(t, saved, updated)
		assert.Equal(t, "curl json", updated.Title().String())
		assert.Equal(t, "POST", updated.Description().String())
		assert.Equal(t, "curl -d\n", updated.FirstFragment().Content().String())
		assert.Equal(t, createdAt(), updated.UpdatedAt())
	})

	t.Run("saves a new Language", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "curl")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		input := updateInput(stored, "curl", "", "curl")
		input.Fragments[0].Language = "Bash"

		updated, err := newUpdate(t, repo).Run(t.Context(), input)

		require.NoError(t, err)
		assert.Equal(t, "Bash", updated.FirstFragment().Language().String())
	})

	t.Run("replaces the Tags with the stored and the new ones it carries", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		http := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "http"})
		shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
		stored := storedSnippet(t, "curl").Retagged([]domain.Tag{http})
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		input := updateInput(stored, "curl", "", "curl")
		input.Tags = []domain.Tag{shell}
		input.NewTags = []string{"api"}

		updated, err := newUpdate(t, repo).Run(t.Context(), input)

		require.NoError(t, err)
		require.Len(t, updated.Tags(), 2)
		assert.Equal(t, "api", updated.Tags()[0].Name().String())
		assert.Equal(t, createdAt(), updated.Tags()[0].CreatedAt())
		assert.Equal(t, shell, updated.Tags()[1])
	})

	t.Run("keeps content with tabs byte for byte", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "if x {\n\treturn\n}\n")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		updated, err := newUpdate(t, repo).Run(
			t.Context(), updateInput(stored, "renamed", "", stored.FirstFragment().Content().String()),
		)

		require.NoError(t, err)
		assert.Equal(t, "if x {\n\treturn\n}\n", updated.FirstFragment().Content().String())
	})

	t.Run("saves the Fragments it lists, in its order and with their titles", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		stored := storedSnippetWith(t, ids, "curl")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		listed := updateInput(stored, "curl", "", "curl")
		listed.Fragments[2].Title = " NOTES "
		added := snippet.FragmentInput{
			ID:       domain.FragmentID{},
			Title:    "put.sh",
			Language: "Bash",
			Content:  "curl -X PUT",
		}
		listed.Fragments = []snippet.FragmentInput{listed.Fragments[2], added, listed.Fragments[0]}

		updated, err := newUpdateWith(t, repo, ids).Run(t.Context(), listed)

		require.NoError(t, err)
		require.Len(t, updated.Fragments(), 3)
		assert.Equal(t, stored.Fragments()[2].ID(), updated.Fragments()[0].ID())
		assert.Equal(t, "NOTES", updated.Fragments()[0].Title().String())
		assert.NotContains(t, []domain.FragmentID{{}, stored.Fragments()[0].ID()}, updated.Fragments()[1].ID())
		assert.Equal(t, "put.sh", updated.Fragments()[1].Title().String())
		assert.Equal(t, "Bash", updated.Fragments()[1].Language().String())
		assert.Equal(t, "curl -X PUT", updated.Fragments()[1].Content().String())
		assert.Equal(t, createdAt(), updated.Fragments()[1].CreatedAt())
		assert.Equal(t, stored.Fragments()[0], updated.Fragments()[2])
	})

	t.Run("passes up a Fragment the Snippet doesn't hold without saving", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "curl")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		input := updateInput(stored, "curl", "", "curl")
		input.Fragments[0].ID = testkit.Fragment(t, testkit.FragmentSpec{}).ID()

		_, err := newUpdate(t, repo).Run(t.Context(), input)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Update: ")
	})

	t.Run("refuses to save no Fragment", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "curl")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		input := updateInput(stored, "curl", "", "curl")
		input.Fragments = nil

		_, err := newUpdate(t, repo).Run(t.Context(), input)

		require.ErrorIs(t, err, domain.ErrNoFragment)
	})

	t.Run("reports every invalid field at once without loading", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")

		input := updateInput(
			stored, " ", strings.Repeat("d", maxDescriptionRunes+1), strings.Repeat("c", maxContentBytes+1),
		)
		input.Fragments[0].Language = "golang"
		input.Fragments[1].Title = strings.Repeat("t", maxFragmentTitleRunes+1)

		_, err := newUpdate(t, NewMockUpdateRepository(t)).Run(t.Context(), input)

		require.ErrorIs(t, err, value.ErrBlankTitle)
		require.ErrorIs(t, err, value.ErrUnknownLanguage)
		require.ErrorIs(t, err, value.ErrFragmentTitleTooLong)
		assert.Equal(
			t,
			[]domain.Field{
				domain.FieldTitle, domain.FieldDescription, domain.FieldLanguage, domain.FieldContent,
				domain.FieldFragmentTitle,
			},
			fieldsOf(domain.FieldErrors(err)),
		)
	})

	t.Run("names the position of the first Fragment with an invalid field", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")
		input := updateInput(stored, "curl", "", "")
		input.Fragments[1].Content = strings.Repeat("c", maxContentBytes+1)

		_, err := newUpdate(t, NewMockUpdateRepository(t)).Run(t.Context(), input)

		fragmentErr, ok := errors.AsType[domain.FragmentError](err)
		require.True(t, ok)
		assert.Equal(t, 1, fragmentErr.Position)
	})

	t.Run("passes up a Snippet that is gone", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Snippet{}, domain.ErrNotFound)

		_, err := newUpdate(t, repo).Run(t.Context(), updateInput(stored, "curl", "", ""))

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Update: ")
	})

	t.Run("passes up a save refused as stale", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")
		repo := NewMockUpdateRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrConflict)

		_, err := newUpdate(t, repo).Run(t.Context(), updateInput(stored, "curl", "", ""))

		require.ErrorIs(t, err, domain.ErrConflict)
		assert.ErrorContains(t, err, "snippet.Update: ")
	})
}

func newUpdate(t *testing.T, repo snippet.UpdateRepository) *snippet.Update {
	t.Helper()

	return newUpdateWith(t, repo, testkit.NewSequentialIDs())
}

func newUpdateWith(t *testing.T, repo snippet.UpdateRepository, ids snippet.IDGenerator) *snippet.Update {
	t.Helper()

	update, err := snippet.NewUpdate(repo, ids, fixedClock())
	require.NoError(t, err)

	return update
}

func storedSnippet(t *testing.T, content string) domain.Snippet {
	t.Helper()

	return storedSnippetWith(t, testkit.NewSequentialIDs(), content)
}

func storedSnippetWith(t *testing.T, ids *testkit.SequentialIDs, content string) domain.Snippet {
	t.Helper()

	created := createdAt().Add(-time.Hour)

	return testkit.Snippet(t, testkit.SnippetSpec{
		Fragments: []testkit.FragmentSpec{
			{ID: ids.NewFragmentID(), Content: content, CreatedAt: created},
			{ID: ids.NewFragmentID(), Title: "body.json", Language: "JSON", Content: "{}", CreatedAt: created},
			{ID: ids.NewFragmentID(), Title: "README", Content: "Posts a body", CreatedAt: created},
		},
		CreatedAt: created,
		UpdatedAt: loadedAt(),
	})
}

func updateInput(stored domain.Snippet, title, description, firstContent string) snippet.UpdateInput {
	fragments := make([]snippet.FragmentInput, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		fragments = append(fragments, snippet.FragmentInput{
			ID:       fragment.ID(),
			Title:    fragment.Title().String(),
			Language: fragment.Language().String(),
			Content:  fragment.Content().String(),
		})
	}

	fragments[0].Content = firstContent

	return snippet.UpdateInput{
		SnippetID:       stored.ID(),
		LoadedUpdatedAt: stored.UpdatedAt(),
		Title:           title,
		Description:     description,
		Fragments:       fragments,
		Tags:            stored.Tags(),
		NewTags:         nil,
	}
}

func loadedAt() time.Time {
	return createdAt().Add(-time.Minute)
}
