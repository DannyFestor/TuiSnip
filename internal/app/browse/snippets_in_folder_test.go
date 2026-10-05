package browse_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

var errDatabaseLocked = errors.New("database is locked")

func TestNewSnippetsInFolder(t *testing.T) {
	t.Parallel()

	_, err := browse.NewSnippetsInFolder(nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	assert.ErrorContains(t, err, "browse.NewSnippetsInFolder: lister")
}

func TestSnippetsInFolder_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns the lister's Snippets for the Folder in the Input's sort order", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		folderID := ids.NewFolderID()
		want := []domain.Snippet{
			testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), Title: "curl json", FolderID: folderID}),
			testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), Title: "git undo", FolderID: folderID}),
		}
		lister := NewMockSnippetLister(t)
		lister.EXPECT().ListInFolder(mock.Anything, folderID, domain.SortOrderCreated).Return(want, nil)

		got, err := newSnippetsInFolder(t, lister).Run(
			t.Context(),
			browse.SnippetsInFolderInput{FolderID: folderID, Order: domain.SortOrderCreated},
		)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("returns the lister's error", func(t *testing.T) {
		t.Parallel()

		lister := NewMockSnippetLister(t)
		lister.EXPECT().
			ListInFolder(mock.Anything, domain.FolderID{}, domain.SortOrderTitle).
			Return(nil, errDatabaseLocked)

		_, err := newSnippetsInFolder(t, lister).Run(
			t.Context(),
			browse.SnippetsInFolderInput{FolderID: domain.FolderID{}, Order: domain.SortOrderTitle},
		)

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.SnippetsInFolder: ")
	})
}

func newSnippetsInFolder(t *testing.T, lister browse.SnippetLister) *browse.SnippetsInFolder {
	t.Helper()

	action, err := browse.NewSnippetsInFolder(lister)
	require.NoError(t, err)

	return action
}
