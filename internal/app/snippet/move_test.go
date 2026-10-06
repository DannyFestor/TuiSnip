package snippet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewMove(t *testing.T) {
	t.Parallel()

	t.Run("names the missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewMove(nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.ErrorContains(t, err, "snippet.NewMove: repo")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewMove(NewMockMoveRepository(t))

		assert.NoError(t, err)
	})
}

func TestMove_Run(t *testing.T) {
	t.Parallel()

	destinations := []struct {
		name     string
		folderID domain.FolderID
	}{
		{name: "into a Folder", folderID: testkit.NewSequentialIDs().NewFolderID()},
		{name: "to the Root", folderID: domain.FolderID{}},
	}
	for _, tt := range destinations {
		t.Run("moves the stored Snippet "+tt.name+", keeping its Language", func(t *testing.T) {
			t.Parallel()

			stored := storedSnippet(t, "curl -d @b.json")
			repo := NewMockMoveRepository(t)
			repo.EXPECT().Find(mock.Anything, stored.ID()).Return(stored, nil)

			var written domain.Snippet

			repo.EXPECT().Move(mock.Anything, mock.AnythingOfType("domain.Snippet")).
				Run(func(_ context.Context, s domain.Snippet) { written = s }).
				Return(nil)

			moved, err := newMove(t, repo).Run(t.Context(), snippet.MoveInput{
				SnippetID: stored.ID(), FolderID: tt.folderID,
			})

			require.NoError(t, err)
			assert.Equal(t, written, moved)
			assert.Equal(t, tt.folderID, moved.FolderID())
			assert.Equal(t, stored.FirstFragment(), moved.FirstFragment())
			assert.Equal(t, stored.UpdatedAt(), moved.UpdatedAt())
		})
	}

	t.Run("passes up a Snippet that is gone", func(t *testing.T) {
		t.Parallel()

		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Snippet{}, domain.ErrNotFound)

		_, err := newMove(t, repo).Run(t.Context(), snippet.MoveInput{
			SnippetID: storedSnippet(t, "").ID(), FolderID: domain.FolderID{},
		})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "snippet.Move: ")
	})

	t.Run("passes up a failed move", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "")
		repo := NewMockMoveRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(stored, nil)
		repo.EXPECT().Move(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newMove(
			t,
			repo,
		).Run(t.Context(), snippet.MoveInput{SnippetID: stored.ID(), FolderID: domain.FolderID{}})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "snippet.Move: ")
	})
}

func newMove(t *testing.T, repo snippet.MoveRepository) *snippet.Move {
	t.Helper()

	move, err := snippet.NewMove(repo)
	require.NoError(t, err)

	return move
}
