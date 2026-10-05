package tui_test

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	editOverlayTitle = "Editing"
	quitQuestion     = "Quit and discard the unsaved changes? [y/N]"
	reloadQuestion   = "This Snippet changed in another TuiSnip. Reload it and discard your changes? [y/N]"
	genericFailure   = "Something went wrong; see the log"
)

func TestModel_editOverlay(t *testing.T) {
	t.Parallel()

	t.Run("saves the Snippet and selects it", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, snippet.CreateInput{
			Title: "Prune", Description: "", Language: "plaintext", Content: "",
		}).Return(snippets[2], nil)
		screen := start(t, creatingModel(t, creator, listerReturning(t, snippets[:2], snippets)), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'))
		screen.press(keypress.Typed("Prune")...)
		screen.press(keypress.Ctrl('s'))

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

		screen.press(keypress.Letter('n'), keypress.Ctrl('s'))

		assert.Contains(t, screen.screen(), "Title is blank")
		assert.Contains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("shows a generic failure when a save fails", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'), keypress.Ctrl('s'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Contains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("updates the selected Snippet and shows the change", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 2)
		edited := testkit.Snippet(t, testkit.SnippetSpec{
			ID:          snippets[0].ID(),
			Title:       "Snippet 1 edited",
			Description: "Description 1",
			Fragment:    testkit.FragmentSpec{ID: snippets[0].FirstFragment().ID()},
		})
		updater := NewMockSnippetUpdater(t)
		updater.EXPECT().Run(mock.Anything, snippet.UpdateInput{
			SnippetID:       snippets[0].ID(),
			LoadedUpdatedAt: snippets[0].UpdatedAt(),
			Title:           "Snippet 1 edited",
			Description:     "Description 1",
			Language:        "plaintext",
			Content:         "",
		}).Return(edited, nil)
		lister := listerReturning(t, snippets, []domain.Snippet{edited, snippets[1]})
		screen := start(t, updatingModel(t, updater, lister), wideWidth, wideHeight)

		screen.press(keypress.Typed("3e edited")...)
		screen.press(keypress.Ctrl('s'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "Snippet 1 edited")
	})

	t.Run("keeps a Tag as the Browse selection after the save", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		stored := filedIn(t, domain.FolderID{})
		edited := renamed(t, stored, filedTitle+" edited")
		updater := NewMockSnippetUpdater(t)
		updater.EXPECT().Run(mock.Anything, mock.Anything).Return(edited, nil)

		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, stored)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, edited)
		screen := start(t, taggedUpdatingModel(t, updater, tagsOf(t, tags...), tagged), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))
		screen.press(keypress.Typed("e edited")...)
		screen.press(keypress.Ctrl('s'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "3 # docker · by title")
		assert.Contains(t, screen.screen(), filedTitle+" edited")
	})

	t.Run("ignores a paste on the main screen", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.send(tea.PasteMsg{Content: "Pasted title"})

		assert.NotContains(t, screen.screen(), "Pasted title")
	})
}

func TestModel_editOverlayStaleSave(t *testing.T) {
	t.Parallel()

	t.Run("reloads the stored Snippet when the reload is accepted", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 2)
		changed := changedElsewhere(t, snippets[0])
		lister := listerReturning(t, snippets, []domain.Snippet{changed, snippets[1]})
		screen := start(t, updatingModel(t, refusingUpdater(t), lister), wideWidth, wideHeight)

		screen.press(keypress.Typed("3e mine")...)
		screen.press(keypress.Ctrl('s'))
		asked := screen.screen()
		screen.press(keypress.Letter('y'))

		assert.Contains(t, asked, reloadQuestion)
		assert.NotContains(t, asked, genericFailure)
		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "Snippet 1 elsewhere")
		assert.NotContains(t, screen.screen(), "Snippet 1 mine")
	})

	t.Run("? does not open help over the Changed elsewhere confirmation", func(t *testing.T) {
		t.Parallel()

		lister := NewMockFolderSnippetsLister(t)
		listingIn(lister, domain.FolderID{}, numberedSnippets(t, 2)...)
		screen := start(t, updatingModel(t, refusingUpdater(t), lister), wideWidth, wideHeight)

		screen.press(keypress.Typed("3e mine")...)
		screen.press(keypress.Ctrl('s'), keypress.Letter('?'))

		assert.Contains(t, screen.screen(), reloadQuestion)
		assert.NotContains(t, screen.screen(), "╭ Help ")
	})

	t.Run("keeps a Tag as the Browse selection after the reload", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		stored := filedIn(t, domain.FolderID{})
		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, stored)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, changedElsewhere(t, stored))
		screen := start(
			t,
			taggedUpdatingModel(t, refusingUpdater(t), tagsOf(t, tags...), tagged),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))
		screen.press(keypress.Typed("e mine")...)
		screen.press(keypress.Ctrl('s'), keypress.Letter('y'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "3 # docker · by title")
		assert.Contains(t, screen.screen(), filedTitle+" elsewhere")
	})

	t.Run("keeps the edits open when the reload is declined", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 2)
		screen := start(t, updatingModel(t, refusingUpdater(t), listerOf(t, snippets...)), wideWidth, wideHeight)

		screen.press(keypress.Typed("3e mine")...)
		screen.press(keypress.Ctrl('s'), keypress.Letter('n'))

		assert.Contains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "Snippet 1 mine")
		assert.NotContains(t, screen.screen(), reloadQuestion)
	})
}

func TestModel_editOverlayForcedQuit(t *testing.T) {
	t.Parallel()

	t.Run("quits on y after asking", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'), keypress.Letter('x'), keypress.Ctrl('c'))
		asked := screen.screen()
		screen.press(keypress.Letter('y'))

		assert.Contains(t, asked, quitQuestion)
		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("quits at once without changes", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'), keypress.Ctrl('c'))

		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("asks before quitting on the configured forced quit key", func(t *testing.T) {
		t.Parallel()

		settings := testsettings.Default(t)
		settings.ForcedQuitKey = "ctrl+q"
		screen := start(t, modelWithSettings(t, actions{
			lister:     listerOf(t),
			treeLister: treeOf(t, emptyTree()),
			copier:     NewMockSnippetCopier(t),
			creator:    NewMockSnippetCreator(t),
			searcher:   NewMockSnippetSearcher(t),
		}, settings), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'), keypress.Letter('z'), keypress.Ctrl('q'))

		assert.Contains(t, screen.screen(), quitQuestion)
	})
}

func updatingModel(t *testing.T, updater *MockSnippetUpdater, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		updater:    updater,
		searcher:   NewMockSnippetSearcher(t),
	})
}

func taggedUpdatingModel(
	t *testing.T, updater *MockSnippetUpdater, tags *MockTagLister, tagged *MockTagSnippetsLister,
) tui.Model {
	t.Helper()

	with := taggedActions(t, listerOf(t), tags, tagged)
	with.updater = updater

	return modelWith(t, with)
}

func refusingUpdater(t *testing.T) *MockSnippetUpdater {
	t.Helper()

	updater := NewMockSnippetUpdater(t)
	updater.EXPECT().
		Run(mock.Anything, mock.Anything).
		Return(domain.Snippet{}, fmt.Errorf("snippet.Update: %w", domain.ErrConflict))

	return updater
}

func changedElsewhere(t *testing.T, loaded domain.Snippet) domain.Snippet {
	t.Helper()

	return renamed(t, loaded, loaded.Title().String()+" elsewhere")
}

func renamed(t *testing.T, loaded domain.Snippet, title string) domain.Snippet {
	t.Helper()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:          loaded.ID(),
		Title:       title,
		Description: loaded.Description().String(),
		Fragment:    testkit.FragmentSpec{ID: loaded.FirstFragment().ID()},
	})
}

func editingModel(t *testing.T, creator *MockSnippetCreator) tui.Model {
	t.Helper()

	return creatingModel(t, creator, listerOf(t))
}

func creatingModel(t *testing.T, creator *MockSnippetCreator, lister *MockFolderSnippetsLister) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    creator,
		searcher:   NewMockSnippetSearcher(t),
	})
}
