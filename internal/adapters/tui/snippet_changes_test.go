package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const snippetDeleteQuestion = `Permanently delete Snippet "Snippet 2"? This cannot be undone. [y/N]`

func TestModel_duplicate(t *testing.T) {
	t.Parallel()

	t.Run("duplicates the selected Snippet and selects the duplicate", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		duplicator := NewMockSnippetDuplicator(t)
		duplicator.EXPECT().Run(mock.Anything, snippet.DuplicateInput{SnippetID: snippets[0].ID()}).
			Return(snippets[2], nil)

		with := snippetChangingActions(t, listerReturning(t, snippets[:2], snippets))
		with.duplicator = duplicator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('c'))

		assert.Contains(t, screen.screen(), "Description 3")
	})

	t.Run("reports a failed duplicate", func(t *testing.T) {
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

func TestModel_snippetDelete(t *testing.T) {
	t.Parallel()

	t.Run("confirms, deletes, and selects the Snippet that takes its row", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		deleter := NewMockSnippetDeleter(t)
		deleter.EXPECT().Run(mock.Anything, snippet.DeleteInput{SnippetID: snippets[1].ID()}).Return(nil)

		with := snippetChangingActions(
			t, listerReturning(t, snippets, []domain.Snippet{snippets[0], snippets[2]}),
		)
		with.deleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)
		screen.press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('d'))

		assert.Contains(t, screen.screen(), snippetDeleteQuestion)

		screen.press(keypress.Letter('y'))

		assert.NotContains(t, screen.screen(), "Snippet 2")
		assert.Contains(t, screen.screen(), "Description 3")
	})

	t.Run("keeps the Snippet when the delete is declined", func(t *testing.T) {
		t.Parallel()

		with := snippetChangingActions(t, listerOf(t, numberedSnippets(t, 3)...))
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('d'), keypress.Letter('n'))

		assert.NotContains(t, screen.screen(), "[y/N]")
		assert.Contains(t, screen.screen(), "Description 2")
	})

	t.Run("reports a failed delete", func(t *testing.T) {
		t.Parallel()

		deleter := NewMockSnippetDeleter(t)
		deleter.EXPECT().Run(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		with := snippetChangingActions(t, listerOf(t, numberedSnippets(t, 3)...))
		with.deleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func snippetChangingActions(t *testing.T, lister tui.FolderSnippetsLister) actions {
	t.Helper()

	return actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
	}
}
