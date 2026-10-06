package folder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewMove(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewMove(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "folder.NewMove: repo")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewMove(NewMockMoveRepository(t))

		assert.NoError(t, err)
	})
}

func TestMove_Run(t *testing.T) {
	t.Parallel()

	t.Run("moves the Folder under another Folder", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		parentID := otherFolderID()
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().ListDescendantIDs(mock.Anything, stored.ID()).Return(nil, nil)
		written := expectMove(repo)

		moved, err := newMove(t, repo).Run(t.Context(), folder.MoveInput{FolderID: stored.ID(), ParentID: parentID})

		require.NoError(t, err)
		assert.Equal(t, *written, moved)
		assert.Equal(t, parentID, moved.ParentID())
		assert.Equal(t, stored.UpdatedAt(), moved.UpdatedAt())
	})

	t.Run("moves the Folder to the Root without looking at its subtree", func(t *testing.T) {
		t.Parallel()

		stored := testkit.Folder(t, testkit.FolderSpec{ID: storedFolder(t).ID(), ParentID: otherFolderID()})
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		written := expectMove(repo)

		moved, err := newMove(
			t,
			repo,
		).Run(t.Context(), folder.MoveInput{FolderID: stored.ID(), ParentID: domain.FolderID{}})

		require.NoError(t, err)
		assert.Equal(t, *written, moved)
		assert.True(t, moved.AtRoot())
	})

	t.Run("refuses a move under the Folder's own descendant", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		descendantID := otherFolderID()
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)
		repo.EXPECT().ListDescendantIDs(mock.Anything, stored.ID()).Return([]domain.FolderID{descendantID}, nil)

		_, err := newMove(t, repo).Run(t.Context(), folder.MoveInput{FolderID: stored.ID(), ParentID: descendantID})

		require.ErrorIs(t, err, domain.ErrFolderCycle)
		assert.ErrorContains(t, err, "folder.Move: ")
	})

	t.Run("passes up a Folder that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Folder{}, domain.ErrNotFound)

		_, err := newMove(
			t,
			repo,
		).Run(t.Context(), folder.MoveInput{FolderID: storedFolder(t).ID(), ParentID: otherFolderID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.Move: ")
	})

	t.Run("passes up a failed descendant listing", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().ListDescendantIDs(mock.Anything, mock.Anything).Return(nil, errDatabaseLocked)

		_, err := newMove(t, repo).Run(t.Context(), folder.MoveInput{FolderID: stored.ID(), ParentID: otherFolderID()})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "folder.Move: ")
	})

	t.Run("passes up a failed move", func(t *testing.T) {
		t.Parallel()

		stored := storedFolder(t)
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().ListDescendantIDs(mock.Anything, mock.Anything).Return(nil, nil)
		repo.EXPECT().Move(mock.Anything, mock.Anything).Return(domain.ErrFolderCycle)

		_, err := newMove(t, repo).Run(t.Context(), folder.MoveInput{FolderID: stored.ID(), ParentID: otherFolderID()})

		require.ErrorIs(t, err, domain.ErrFolderCycle)
		assert.ErrorContains(t, err, "folder.Move: ")
	})
}

func newMove(t *testing.T, repo folder.MoveRepository) *folder.Move {
	t.Helper()

	move, err := folder.NewMove(repo)
	require.NoError(t, err)

	return move
}

func expectMove(repo *MockMoveRepository) *domain.Folder {
	var written domain.Folder

	repo.EXPECT().Move(mock.Anything, mock.AnythingOfType("domain.Folder")).
		Run(func(_ context.Context, f domain.Folder) { written = f }).
		Return(nil)

	return &written
}

func otherFolderID() domain.FolderID {
	ids := testkit.NewSequentialIDs()
	ids.NewFolderID()

	return ids.NewFolderID()
}
