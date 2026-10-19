package tui_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_goneElsewhere(t *testing.T) {
	t.Parallel()

	t.Run("names the Snippet another TuiSnip took away and reloads the list without it", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		duplicator := NewMockSnippetDuplicator(t)
		duplicator.EXPECT().Run(mock.Anything, mock.Anything).
			Return(domain.Snippet{}, fmt.Errorf("snippet.Duplicate: %w", domain.ErrNotFound))

		with := snippetChangingActions(t, listerReturning(t, snippets, []domain.Snippet{snippets[0], snippets[2]}))
		with.duplicator = duplicator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('c'))

		assert.Contains(t, screen.screen(),
			`Snippet "Snippet 2" was moved to the Trash or deleted in another TuiSnip`)
		assert.NotContains(t, screen.screen(), "Description 2")
	})

	t.Run("names the Folder a moved Snippet was going into", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		snippets := sampleSnippets(t)
		mover := NewMockSnippetMover(t)
		mover.EXPECT().Run(mock.Anything, mock.Anything).
			Return(domain.Snippet{}, fmt.Errorf("snippet.Move: %w", domain.ErrFolderGone))

		with := browsingActions(t, sample, listerReturning(t, snippets, snippets))
		with.mover = mover
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'))
		pickedFolder(screen, "docker")

		assert.Contains(t, screen.screen(), `Folder "docker" was moved to the Trash or deleted in another TuiSnip`)
	})

	t.Run("keeps the generic failure for anything else", func(t *testing.T) {
		t.Parallel()

		duplicator := NewMockSnippetDuplicator(t)
		duplicator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, errDatabaseLocked)

		with := snippetChangingActions(t, listerOf(t, numberedSnippets(t, 1)...))
		with.duplicator = duplicator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('c'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}
