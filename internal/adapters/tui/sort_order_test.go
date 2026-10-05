package tui_test

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
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
		settings.Remembered.SortOrder = domain.SortOrderCreated

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

func TestModel_rejectedSortOrder(t *testing.T) {
	t.Parallel()

	t.Run("refuses to start with an unknown order", func(t *testing.T) {
		t.Parallel()

		settings := testsettings.Default(t)
		settings.Remembered.SortOrder = domain.SortOrder("language")

		_, err := tui.New(t.Context(), tui.Deps{
			Lister:                NewMockFolderSnippetsLister(t),
			TreeLister:            NewMockFolderTreeLister(t),
			TagLister:             NewMockTagLister(t),
			TagSnippetsLister:     NewMockTagSnippetsLister(t),
			Copier:                NewMockSnippetCopier(t),
			Creator:               NewMockSnippetCreator(t),
			Capturer:              NewMockSnippetCapturer(t),
			Updater:               NewMockSnippetUpdater(t),
			Searcher:              NewMockSnippetSearcher(t),
			FolderCreator:         NewMockFolderCreator(t),
			FolderRenamer:         NewMockFolderRenamer(t),
			FolderDeletePreviewer: NewMockFolderDeletePreviewer(t),
			FolderDeleter:         NewMockFolderDeleter(t),
			TagCreator:            NewMockTagCreator(t),
			TagRenamer:            NewMockTagRenamer(t),
			TagDeletePreviewer:    NewMockTagDeletePreviewer(t),
			TagDeleter:            NewMockTagDeleter(t),
			SortOrderSaver:        NewMockSortOrderSaver(t),
			CollapsedFoldersSaver: NewMockCollapsedFoldersSaver(t),
			Settings:              settings,
			Logger:                slog.New(slog.DiscardHandler),
		})

		require.ErrorIs(t, err, domain.ErrInvalidSortOrder)
	})

	t.Run("shows the failure when Snippets arrive in an unknown order", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := start(t, newModel(t, listerOf(t, snippets...), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.send(mainscreen.SnippetsLoaded{Snippets: snippets, Order: domain.SortOrder("language")})

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Contains(t, screen.screen(), "3 Root · by title")
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
