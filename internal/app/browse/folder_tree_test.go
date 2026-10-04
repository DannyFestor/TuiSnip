package browse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewFolderTree(t *testing.T) {
	t.Parallel()

	_, err := browse.NewFolderTree(nil, nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "browse.NewFolderTree: folders")
	assert.ErrorContains(t, err, "counter")
}

func TestFolderTree_Run(t *testing.T) {
	t.Parallel()

	t.Run("nests each Folder under its parent, alphabetically, with its Snippet count", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		golang := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
		docker := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "Docker"})
		tests := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "testing", ParentID: golang.ID()})
		errs := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "errors", ParentID: golang.ID()})
		folders := folderListerOf(t, golang, docker, tests, errs)
		counter := snippetCounterOf(t, map[domain.FolderID]int{{}: 3, golang.ID(): 2, tests.ID(): 1})

		got, err := newFolderTree(t, folders, counter).Run(t.Context(), browse.FolderTreeInput{})

		require.NoError(t, err)
		assert.Equal(t, browse.Tree{
			RootSnippetCount: 3,
			Folders: []browse.FolderNode{
				{Folder: docker, SnippetCount: 0, Children: nil},
				{Folder: golang, SnippetCount: 2, Children: []browse.FolderNode{
					{Folder: errs, SnippetCount: 0, Children: nil},
					{Folder: tests, SnippetCount: 1, Children: nil},
				}},
			},
		}, got)
	})

	t.Run("returns only the Root when there are no Folders", func(t *testing.T) {
		t.Parallel()

		got, err := newFolderTree(t, folderListerOf(t), snippetCounterOf(t, nil)).
			Run(t.Context(), browse.FolderTreeInput{})

		require.NoError(t, err)
		assert.Equal(t, browse.Tree{RootSnippetCount: 0, Folders: nil}, got)
	})

	t.Run("returns the lister's error", func(t *testing.T) {
		t.Parallel()

		folders := NewMockFolderLister(t)
		folders.EXPECT().List(mock.Anything).Return(nil, errDatabaseLocked)

		_, err := newFolderTree(t, folders, NewMockSnippetCounter(t)).Run(t.Context(), browse.FolderTreeInput{})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.FolderTree: ")
	})

	t.Run("returns the counter's error", func(t *testing.T) {
		t.Parallel()

		counter := NewMockSnippetCounter(t)
		counter.EXPECT().CountByFolder(mock.Anything).Return(nil, errDatabaseLocked)

		_, err := newFolderTree(t, folderListerOf(t), counter).Run(t.Context(), browse.FolderTreeInput{})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "browse.FolderTree: ")
	})
}

func newFolderTree(t *testing.T, folders browse.FolderLister, counter browse.SnippetCounter) *browse.FolderTree {
	t.Helper()

	action, err := browse.NewFolderTree(folders, counter)
	require.NoError(t, err)

	return action
}

func folderListerOf(t *testing.T, folders ...domain.Folder) *MockFolderLister {
	t.Helper()

	lister := NewMockFolderLister(t)
	lister.EXPECT().List(mock.Anything).Return(folders, nil)

	return lister
}

func snippetCounterOf(t *testing.T, counts map[domain.FolderID]int) *MockSnippetCounter {
	t.Helper()

	counter := NewMockSnippetCounter(t)
	counter.EXPECT().CountByFolder(mock.Anything).Return(counts, nil)

	return counter
}
