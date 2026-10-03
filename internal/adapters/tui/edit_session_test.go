package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	editOverlayTitle = "Editing"
	quitQuestion     = "Quit and discard the unsaved changes? y/N"
)

func TestModel_editOverlay(t *testing.T) {
	t.Parallel()

	t.Run("saves the Snippet and selects it", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, snippet.CreateInput{Title: "Prune"}).Return(snippets[2], nil)
		screen := start(t, creatingModel(t, creator, listerReturning(t, snippets[:2], snippets)), wideWidth, wideHeight)

		screen.press(letter('n'))
		screen.press(typed("Prune")...)
		screen.press(ctrl('s'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "Description 3")
	})

	t.Run("shows why a save was rejected", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, mock.Anything).
			Return(domain.Snippet{}, domain.OnField(domain.FieldTitle, value.ErrBlankTitle))
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('s'))

		assert.Contains(t, screen.screen(), "Title is blank")
		assert.Contains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("shows a generic failure when a save fails", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('s'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Contains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("ignores a paste on the main screen", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.send(tea.PasteMsg{Content: "Pasted title"})

		assert.NotContains(t, screen.screen(), "Pasted title")
	})
}

func TestModel_editOverlayForcedQuit(t *testing.T) {
	t.Parallel()

	t.Run("quits on y after asking", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'), ctrl('c'))
		asked := screen.screen()
		screen.press(letter('y'))

		assert.Contains(t, asked, quitQuestion)
		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("quits at once without changes", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('c'))

		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("asks before quitting on the configured forced quit key", func(t *testing.T) {
		t.Parallel()

		settings := testsettings.Default(t)
		settings.ForcedQuitKey = "ctrl+q"
		screen := start(t, modelWithSettings(t, actions{
			lister:   listerOf(t),
			copier:   NewMockSnippetCopier(t),
			creator:  NewMockSnippetCreator(t),
			searcher: NewMockSnippetSearcher(t),
		}, settings), wideWidth, wideHeight)

		screen.press(letter('n'), letter('z'), ctrl('q'))

		assert.Contains(t, screen.screen(), quitQuestion)
	})
}

func editingModel(t *testing.T, creator *MockSnippetCreator) tui.Model {
	t.Helper()

	return creatingModel(t, creator, listerOf(t))
}

func creatingModel(t *testing.T, creator *MockSnippetCreator, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:   lister,
		copier:   NewMockSnippetCopier(t),
		creator:  creator,
		searcher: NewMockSnippetSearcher(t),
	})
}
