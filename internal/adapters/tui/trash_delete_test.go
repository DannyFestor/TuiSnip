package tui_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_trashDeletePermanently(t *testing.T) {
	t.Parallel()

	t.Run("deletes the Trash item for good once confirmed and reloads the Trash", func(t *testing.T) {
		t.Parallel()

		prune := trashedPrune(t)
		deleter := NewMockTrashDeleter(t)
		deleter.EXPECT().Run(mock.Anything, trash.DeletePermanentlyInput{Item: prune.ID()}).Return(nil)

		with := trashActions(t, listingTrashThen(t, []trash.Item{prune}, nil))
		with.trashDeleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "The Trash is empty.")
	})

	t.Run("says a Trash item restored elsewhere is no longer in the Trash and reloads", func(t *testing.T) {
		t.Parallel()

		deleter := NewMockTrashDeleter(t)
		deleter.EXPECT().Run(mock.Anything, mock.Anything).
			Return(fmt.Errorf("trash.DeletePermanently: %w", domain.ErrNotFound))

		with := trashActions(t, listingTrashThen(t, []trash.Item{trashedPrune(t)}, nil))
		with.trashDeleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), `"Prune everything" is no longer in the Trash`)
		assert.Contains(t, screen.screen(), "The Trash is empty.")
	})
}

func TestModel_trashEmpty(t *testing.T) {
	t.Parallel()

	emptyings := []struct {
		name    string
		emptied domain.TrashEmptied
		want    string
	}{
		{
			name:    "says how many Folders and Snippets emptying the Trash deleted",
			emptied: domain.TrashEmptied{Folders: 2, Snippets: 20},
			want:    "Emptied the Trash: 2 Folders, 20 Snippets",
		},
		{
			name:    "leaves a zero count out and says 1 for one",
			emptied: domain.TrashEmptied{Folders: 0, Snippets: 1},
			want:    "Emptied the Trash: 1 Snippet",
		},
	}
	for _, tt := range emptyings {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prune, docker := trashedPrune(t), trashedDocker(t)
			emptier := NewMockTrashEmptier(t)
			emptier.EXPECT().Run(mock.Anything, trash.EmptyInput{Items: []trash.ItemID{prune.ID(), docker.ID()}}).
				Return(tt.emptied, nil)

			with := trashActions(t, listingTrashThen(t, []trash.Item{prune, docker}, nil))
			with.trashEmptier = emptier
			screen := start(t, modelWith(t, with), wideWidth, wideHeight)

			screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('D'), keypress.Letter('y'))

			assert.Contains(t, screen.screen(), tt.want)
			assert.Contains(t, screen.screen(), "The Trash is empty.")
		})
	}

	t.Run("says the Trash was emptied when everything counted had left it, and lists what is left", func(t *testing.T) {
		t.Parallel()

		emptier := NewMockTrashEmptier(t)
		emptier.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.TrashEmptied{Folders: 0, Snippets: 0}, nil)

		with := trashActions(t, listingTrashThen(t, []trash.Item{trashedPrune(t)}, []trash.Item{trashedDocker(t)}))
		with.trashEmptier = emptier
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('G'), keypress.Letter('3'), keypress.Letter('D'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "Emptied the Trash")
		assert.NotContains(t, screen.screen(), "Emptied the Trash:")
		assert.Contains(t, screen.screen(), "▸ docker")
	})
}
