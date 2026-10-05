package tui_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestModel_sortOrder(t *testing.T) {
	t.Parallel()

	t.Run("lists in the remembered order from the start", func(t *testing.T) {
		t.Parallel()

		lister := NewMockFolderSnippetsLister(t)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderCreated, sampleSnippets(t)...)
		settings := testsettings.Default(t)
		settings.SortOrder = domain.SortOrderCreated

		screen := start(t, modelWithSettings(t, sorting(t, lister, nil), settings), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "3 Root · by creation date")
	})

	t.Run("s remembers the next order and relists in it, keeping the selected Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		lister := NewMockFolderSnippetsLister(t)
		listingIn(lister, domain.FolderID{}, snippets...)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderUpdated, reversed(snippets)...)

		saver := NewMockSortOrderSaver(t)
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderUpdated).Return(nil).Once()
		screen := start(t, modelWith(t, sorting(t, lister, saver)), wideWidth, wideHeight)

		screen.press(keypress.Typed("3s")...)

		assert.Contains(t, screen.screen(), "3 Root · by last updated")
		assert.Contains(t, screen.screen(), "Stop accepting, drain, exit")
	})

	t.Run("s again moves on to creation date", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		lister := NewMockFolderSnippetsLister(t)
		listingIn(lister, domain.FolderID{}, snippets...)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderUpdated, snippets...)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderCreated, snippets...)

		saver := NewMockSortOrderSaver(t)
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderUpdated).Return(nil).Once()
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderCreated).Return(nil).Once()
		screen := start(t, modelWith(t, sorting(t, lister, saver)), wideWidth, wideHeight)

		screen.press(keypress.Typed("3ss")...)

		assert.Contains(t, screen.screen(), "3 Root · by creation date")
	})

	t.Run("a failed save shows the failure and still relists in the next order", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		lister := NewMockFolderSnippetsLister(t)
		listingIn(lister, domain.FolderID{}, snippets...)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderUpdated, snippets...)

		saver := NewMockSortOrderSaver(t)
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderUpdated).Return(errDatabaseLocked).Once()
		screen := start(t, modelWith(t, sorting(t, lister, saver)), wideWidth, wideHeight)

		screen.press(keypress.Typed("3s")...)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Contains(t, screen.screen(), "3 Root · by last updated")
	})

	t.Run("Search with an empty query lists in the current order", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		lister := NewMockFolderSnippetsLister(t)
		listingIn(lister, domain.FolderID{}, snippets...)
		listingInOrder(lister, domain.FolderID{}, domain.SortOrderUpdated, reversed(snippets)...)

		saver := NewMockSortOrderSaver(t)
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderUpdated).Return(nil).Once()
		screen := start(t, modelWith(t, sorting(t, lister, saver)), wideWidth, wideHeight)
		screen.press(keypress.Typed("3s")...)

		screen.press(keypress.Letter('/'))

		assert.Contains(t, screen.screen(), searchTitle+"2 results")
		assert.Contains(t, screen.screen(), "Reclaim disk space", "previews the first Snippet in the current order")
	})
}

func sorting(t *testing.T, lister *MockFolderSnippetsLister, saver *MockSortOrderSaver) actions {
	t.Helper()

	with := actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
	}
	if saver != nil {
		with.sortOrderSaver = saver
	}

	return with
}

func reversed(snippets []domain.Snippet) []domain.Snippet {
	backwards := slices.Clone(snippets)
	slices.Reverse(backwards)

	return backwards
}
