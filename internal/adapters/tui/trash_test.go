package tui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_trash(t *testing.T) {
	t.Parallel()

	t.Run("lists the Trash once it becomes the Browse selection", func(t *testing.T) {
		t.Parallel()

		with := trashActions(t, listingTrash(t, trashedPrune(t)))
		screen := start(t, modelWith(t, with), 2*wideWidth, wideHeight)

		screen.press(keypress.Letter('G'))

		assert.Contains(t, screen.screen(), "3 Trash")
		assert.Contains(t, screen.screen(), "Root / go / testing · Bash")
		assert.Contains(t, screen.screen(), "deleted 2026-10-14")
	})

	restorations := []struct {
		name        string
		item        func(t *testing.T) trash.Item
		restoration domain.Restoration
		want        string
	}{
		{
			name: "says where a restored item went",
			item: trashedPrune,
			restoration: domain.Restoration{
				FolderID: foldertree.New(t).Tests.ID(),
				Fallback: domain.RestoreFallbackNone,
			},
			want: `Restored "Prune everything" to go / testing`,
		},
		{
			name:        "says an item restored to the Root went to the Root",
			item:        trashedPrune,
			restoration: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackNone},
			want:        `Restored "Prune everything" to the Root`,
		},
		{
			name:        "says a restored Snippet whose Folder is in the Trash went to the Root",
			item:        trashedPrune,
			restoration: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderInTrash},
			want:        `Restored "Prune everything" to the Root; its Folder is in the Trash`,
		},
		{
			name:        "says a restored Snippet whose Folder was deleted went to the Root",
			item:        trashedPrune,
			restoration: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderDeleted},
			want:        `Restored "Prune everything" to the Root; its Folder was deleted`,
		},
		{
			name:        "says a restored Folder whose parent is in the Trash went to the Root",
			item:        trashedDocker,
			restoration: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderInTrash},
			want:        `Restored "docker" to the Root; its parent Folder is in the Trash`,
		},
		{
			name:        "says a restored Folder whose parent was deleted went to the Root",
			item:        trashedDocker,
			restoration: domain.Restoration{FolderID: domain.FolderID{}, Fallback: domain.RestoreFallbackFolderDeleted},
			want:        `Restored "docker" to the Root; its parent Folder was deleted`,
		},
	}
	for _, tt := range restorations {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item := tt.item(t)
			restorer := NewMockTrashRestorer(t)
			restorer.EXPECT().Run(mock.Anything, trash.RestoreInput{Item: item.ID()}).Return(tt.restoration, nil)

			with := trashActions(t, listingTrashThen(t, []trash.Item{item}, nil))
			with.trashRestorer = restorer
			screen := start(t, modelWith(t, with), wideWidth, wideHeight)

			screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('r'))

			assert.Contains(t, screen.screen(), tt.want)
			assert.Contains(t, screen.screen(), "The Trash is empty.")
		})
	}

	t.Run("says an item restored elsewhere is no longer in the Trash and reloads", func(t *testing.T) {
		t.Parallel()

		prune := trashedPrune(t)
		restorer := NewMockTrashRestorer(t)
		restorer.EXPECT().Run(mock.Anything, mock.Anything).
			Return(domain.Restoration{}, fmt.Errorf("trash.Restore: %w", domain.ErrNotFound))

		with := trashActions(t, listingTrashThen(t, []trash.Item{prune}, nil))
		with.trashRestorer = restorer
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('r'))

		assert.Contains(t, screen.screen(), `"Prune everything" is no longer in the Trash`)
		assert.Contains(t, screen.screen(), "The Trash is empty.")
	})

	t.Run("copies a Snippet Trash item through the Trash", func(t *testing.T) {
		t.Parallel()

		prune := trashedPrune(t)
		copier := NewMockSnippetCopier(t)
		copier.EXPECT().Run(mock.Anything, snippet.CopyInput{
			SnippetID: prune.Snippet.ID(), FragmentID: prune.Snippet.FirstFragment().ID(),
		}).Return(snippet.CopyResult{Delivery: domain.CopyDeliveryPlaced}, nil)

		with := trashActions(t, listingTrash(t, prune))
		with.trashCopier = copier
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "Copied")
	})

	t.Run("says a copied item restored elsewhere is no longer in the Trash", func(t *testing.T) {
		t.Parallel()

		prune := trashedPrune(t)
		copier := NewMockSnippetCopier(t)
		copier.EXPECT().Run(mock.Anything, mock.Anything).
			Return(snippet.CopyResult{}, fmt.Errorf("snippet.Copy: %w", domain.ErrNotFound))

		with := trashActions(t, listingTrashThen(t, []trash.Item{prune}, nil))
		with.trashCopier = copier
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), `"Prune everything" is no longer in the Trash`)
	})
}

func trashActions(t *testing.T, lister *MockTrashLister) actions {
	t.Helper()

	with := browsingActions(t, foldertree.New(t), listerOf(t))
	with.trashLister = lister

	return with
}

func listingTrash(t *testing.T, items ...trash.Item) *MockTrashLister {
	t.Helper()

	lister := NewMockTrashLister(t)
	lister.EXPECT().Run(mock.Anything, trash.ListInput{}).Return(items, nil)

	return lister
}

func listingTrashThen(t *testing.T, first, then []trash.Item) *MockTrashLister {
	t.Helper()

	lister := NewMockTrashLister(t)
	lister.EXPECT().Run(mock.Anything, trash.ListInput{}).Return(first, nil).Once()
	lister.EXPECT().Run(mock.Anything, trash.ListInput{}).Return(then, nil)

	return lister
}

func trashedPrune(t *testing.T) trash.Item {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return trash.Item{
		Kind: trash.ItemKindSnippet,
		Snippet: testkit.Snippet(t, testkit.SnippetSpec{
			ID:    ids.NewSnippetID(),
			Title: "Prune everything",
			Fragments: []testkit.FragmentSpec{
				{ID: ids.NewFragmentID(), Language: "Bash", Content: "docker system prune"},
			},
		}),
		DeletedAt: time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC),
		Path:      []string{"go", "testing"},
	}
}

func trashedDocker(t *testing.T) trash.Item {
	t.Helper()

	return trash.Item{
		Kind:      trash.ItemKindFolder,
		Folder:    testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "docker"}),
		DeletedAt: time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC),
		Path:      []string{"ops"},
	}
}
