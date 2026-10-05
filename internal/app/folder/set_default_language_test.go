package folder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestNewSetDefaultLanguage(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewSetDefaultLanguage(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "folder.NewSetDefaultLanguage: repo")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewSetDefaultLanguage(NewMockEditRepository(t), fixedClock())

		assert.NoError(t, err)
	})
}

func TestSetDefaultLanguage_Run(t *testing.T) {
	t.Parallel()

	t.Run("updates the Folder with the Default Language and the time", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockEditRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		var updated domain.Folder

		repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("domain.Folder")).
			Run(func(_ context.Context, f domain.Folder) { updated = f }).
			Return(nil)

		changed, err := newSetDefaultLanguage(t, repo).Run(
			t.Context(), folder.SetDefaultLanguageInput{FolderID: stored.ID(), Language: "Go"},
		)

		require.NoError(t, err)
		assert.Equal(t, updated, changed)
		assert.Equal(t, "Go", changed.DefaultLanguage().String())
		assert.Equal(t, now(), changed.UpdatedAt())
		assert.Equal(t, stored.Name(), changed.Name())
	})

	t.Run("refuses an unknown Language on the language field without loading", func(t *testing.T) {
		t.Parallel()

		_, err := newSetDefaultLanguage(t, NewMockEditRepository(t)).Run(
			t.Context(), folder.SetDefaultLanguageInput{FolderID: storedFolder(t).ID(), Language: "golang"},
		)

		require.ErrorIs(t, err, value.ErrUnknownLanguage)
		fieldErr, ok := errors.AsType[domain.FieldError](err)
		require.True(t, ok)
		assert.Equal(t, domain.FieldLanguage, fieldErr.Field)
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockEditRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Folder{}, domain.ErrNotFound)

		_, err := newSetDefaultLanguage(t, repo).Run(
			t.Context(), folder.SetDefaultLanguageInput{FolderID: storedFolder(t).ID(), Language: "Go"},
		)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.SetDefaultLanguage")
	})

	t.Run("passes up a failed update", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockEditRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newSetDefaultLanguage(t, repo).Run(
			t.Context(), folder.SetDefaultLanguageInput{FolderID: stored.ID(), Language: "Go"},
		)

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "folder.SetDefaultLanguage")
	})
}

func newSetDefaultLanguage(t *testing.T, repo folder.EditRepository) *folder.SetDefaultLanguage {
	t.Helper()

	setDefaultLanguage, err := folder.NewSetDefaultLanguage(repo, fixedClock())
	require.NoError(t, err)

	return setDefaultLanguage
}
