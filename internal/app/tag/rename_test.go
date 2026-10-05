package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewRename(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewRename(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "tag.NewRename: repo")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewRename(NewMockRenameRepository(t), fixedClock())

		assert.NoError(t, err)
	})
}

func TestRename_Run(t *testing.T) {
	t.Parallel()

	t.Run("renames the Tag with the trimmed name and the time", func(t *testing.T) {
		t.Parallel()

		stored := storedTag(t)
		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		var renamed domain.Tag

		repo.EXPECT().Rename(mock.Anything, mock.AnythingOfType("domain.Tag")).
			RunAndReturn(func(_ context.Context, asked domain.Tag) (domain.Tag, error) {
				renamed = asked

				return asked, nil
			})

		_, err := newRename(t, repo).Run(t.Context(), tag.RenameInput{TagID: stored.ID(), Name: " golang "})

		require.NoError(t, err)
		assert.Equal(t, stored.ID(), renamed.ID())
		assert.Equal(t, "golang", renamed.Name().String())
		assert.Equal(t, now(), renamed.UpdatedAt())
	})

	t.Run("returns the Tag that survives a merge", func(t *testing.T) {
		t.Parallel()

		stored := storedTag(t)
		survivor := testkit.Tag(t, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "Go"})
		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().Rename(mock.Anything, mock.Anything).Return(survivor, nil)

		got, err := newRename(t, repo).Run(t.Context(), tag.RenameInput{TagID: stored.ID(), Name: "Go"})

		require.NoError(t, err)
		assert.Equal(t, survivor, got)
	})

	t.Run("refuses a blank name on the tag_name field", func(t *testing.T) {
		t.Parallel()

		_, err := newRename(t, NewMockRenameRepository(t)).Run(t.Context(), tag.RenameInput{
			TagID: storedTag(t).ID(), Name: "  ",
		})

		require.ErrorIs(t, err, value.ErrBlankTagName)
		fieldErr, ok := errors.AsType[domain.FieldError](err)
		require.True(t, ok)
		assert.Equal(t, domain.FieldTagName, fieldErr.Field)
	})

	t.Run("passes up a Tag that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Tag{}, domain.ErrNotFound)

		_, err := newRename(t, repo).Run(t.Context(), tag.RenameInput{TagID: storedTag(t).ID(), Name: "golang"})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "tag.Rename")
	})

	t.Run("passes up a failed rename", func(t *testing.T) {
		t.Parallel()

		stored := storedTag(t)
		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Rename(mock.Anything, mock.Anything).Return(domain.Tag{}, errDatabaseLocked)

		_, err := newRename(t, repo).Run(t.Context(), tag.RenameInput{TagID: stored.ID(), Name: "golang"})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "tag.Rename")
	})
}

func newRename(t *testing.T, repo tag.RenameRepository) *tag.Rename {
	t.Helper()

	rename, err := tag.NewRename(repo, fixedClock())
	require.NoError(t, err)

	return rename
}
