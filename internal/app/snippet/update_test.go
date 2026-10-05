package snippet_test

import (
	"context"
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

		_, err := snippet.NewUpdate(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "snippet.NewUpdate: repo")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewUpdate(NewMockUpdateRepository(t), fixedClock())

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

	t.Run("reports every invalid field at once without loading", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")

		_, err := newUpdate(t, NewMockUpdateRepository(t)).Run(t.Context(), updateInput(
			stored, " ", strings.Repeat("d", maxDescriptionRunes+1), strings.Repeat("c", maxContentBytes+1),
		))

		require.ErrorIs(t, err, value.ErrBlankTitle)
		assert.Equal(
			t,
			[]domain.Field{domain.FieldTitle, domain.FieldDescription, domain.FieldContent},
			fieldsOf(domain.FieldErrors(err)),
		)
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

	update, err := snippet.NewUpdate(repo, fixedClock())
	require.NoError(t, err)

	return update
}

func storedSnippet(t *testing.T, content string) domain.Snippet {
	t.Helper()

	created := createdAt().Add(-time.Hour)

	return testkit.Snippet(t, testkit.SnippetSpec{
		Fragment:  testkit.FragmentSpec{Content: content, CreatedAt: created},
		CreatedAt: created,
		UpdatedAt: loadedAt(),
	})
}

func updateInput(stored domain.Snippet, title, description, content string) snippet.UpdateInput {
	return snippet.UpdateInput{
		SnippetID:       stored.ID(),
		LoadedUpdatedAt: stored.UpdatedAt(),
		Title:           title,
		Description:     description,
		Content:         content,
	}
}

func loadedAt() time.Time {
	return createdAt().Add(-time.Minute)
}
