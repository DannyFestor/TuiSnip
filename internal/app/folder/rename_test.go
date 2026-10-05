package folder_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewRename(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewRename(nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "folder.NewRename: repo")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewRename(NewMockRenameRepository(t), fixedClock())

		assert.NoError(t, err)
	})
}

func TestRename_Run(t *testing.T) {
	t.Parallel()

	t.Run("updates the Folder with the trimmed name and the time", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

		var updated domain.Folder

		repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("domain.Folder")).
			Run(func(_ context.Context, f domain.Folder) { updated = f }).
			Return(nil)

		renamed, err := newRename(t, repo).Run(t.Context(), folder.RenameInput{FolderID: stored.ID(), Name: " golang "})

		require.NoError(t, err)
		assert.Equal(t, updated, renamed)
		assert.Equal(t, "golang", renamed.Name().String())
		assert.Equal(t, now(), renamed.UpdatedAt())
		assert.Equal(t, stored.ID(), renamed.ID())
	})

	t.Run("refuses a name over 200 characters on the folder_name field", func(t *testing.T) {
		t.Parallel()

		_, err := newRename(t, NewMockRenameRepository(t)).Run(t.Context(), folder.RenameInput{
			FolderID: storedFolder(t).ID(), Name: longName(),
		})

		require.ErrorIs(t, err, value.ErrFolderNameTooLong)
		fieldErr, ok := errors.AsType[domain.FieldError](err)
		require.True(t, ok)
		assert.Equal(t, domain.FieldFolderName, fieldErr.Field)
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Folder{}, domain.ErrNotFound)

		_, err := newRename(
			t,
			repo,
		).Run(t.Context(), folder.RenameInput{FolderID: storedFolder(t).ID(), Name: "golang"})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.Rename")
	})

	t.Run("passes up a failed update", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockRenameRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Update(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newRename(t, repo).Run(t.Context(), folder.RenameInput{FolderID: stored.ID(), Name: "golang"})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "folder.Rename")
	})
}

func newRename(t *testing.T, repo folder.RenameRepository) *folder.Rename {
	t.Helper()

	rename, err := folder.NewRename(repo, fixedClock())
	require.NoError(t, err)

	return rename
}

func storedFolder(t *testing.T) domain.Folder {
	t.Helper()

	return testkit.Folder(t, testkit.FolderSpec{
		ID:        testkit.NewSequentialIDs().NewFolderID(),
		Name:      "go",
		CreatedAt: now().Add(-time.Hour),
	})
}

func longName() string {
	const overMaxRunes = 201

	runes := make([]rune, overMaxRunes)
	for index := range runes {
		runes[index] = 'é'
	}

	return string(runes)
}
