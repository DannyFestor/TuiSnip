package tui_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	editOverlayTitle = "Editing"
	unsavedTitle     = "Editing •"
	contentEntryHint = "enter or down to edit"
	discardQuestion  = "Discard the unsaved changes? y/N"
	quitQuestion     = "Quit and discard the unsaved changes? y/N"
)

func TestModel_editOverlay(t *testing.T) {
	t.Parallel()

	t.Run("n opens the overlay on Title", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'))

		assert.Contains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "› Title")
		assert.Contains(t, screen.screen(), contentEntryHint)
		assert.Contains(t, screen.screen(), "ctrl+s save · esc cancel · down field")
	})

	t.Run("marks unsaved changes in the title", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'))

		assert.Contains(t, screen.screen(), unsavedTitle)
	})

	t.Run("types Pane keys into the field instead of acting on them", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'))
		screen.press(typed("q4/")...)

		assert.NotContains(t, screen.emitted, tea.QuitMsg{})
		assert.Contains(t, screen.screen(), "q4/")
	})

	t.Run("shows the content Bindings inside Content", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(enterContent()...)

		assert.Contains(t, screen.screen(), "ctrl+s save · esc leave")
		assert.NotContains(t, screen.screen(), contentEntryHint)
	})
}

func TestModel_editOverlaySave(t *testing.T) {
	t.Parallel()

	t.Run("creates the Snippet from every field and selects it", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 3)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, snippet.CreateInput{
			Title:       "Prune",
			Description: "Reclaim",
			Content:     "docker\nprune",
		}).Return(snippets[2], nil)
		screen := start(t, modelWith(t, actions{
			lister:   listerReturning(t, snippets[:2], snippets),
			copier:   NewMockSnippetCopier(t),
			creator:  creator,
			searcher: NewMockSnippetSearcher(t),
		}), wideWidth, wideHeight)

		screen.press(letter('n'))
		screen.press(typed("Prune")...)
		screen.press(special(tea.KeyTab))
		screen.press(typed("Reclaim")...)
		screen.press(special(tea.KeyEnter), special(tea.KeyEnter))
		screen.press(typed("docker")...)
		screen.press(special(tea.KeyEnter))
		screen.press(typed("prune")...)
		screen.press(ctrl('s'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
		assert.Contains(t, screen.screen(), "Description 3")
	})

	t.Run("up on the first line of Content returns to Description", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "", Description: "more", Content: "body"}).
			Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), special(tea.KeyDown), special(tea.KeyDown), special(tea.KeyDown))
		screen.press(typed("body")...)
		screen.press(special(tea.KeyUp))
		screen.press(typed("more")...)
		screen.press(ctrl('s'))

		assert.Contains(t, screen.screen(), contentEntryHint)
	})

	t.Run("up below the first line of Content stays in Content", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "", Description: "", Content: "one more\ntwo"}).
			Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(enterContent()...)
		screen.press(typed("one")...)
		screen.press(special(tea.KeyEnter))
		screen.press(typed("two")...)
		screen.press(special(tea.KeyUp))
		screen.press(typed(" more")...)
		screen.press(ctrl('s'))
	})

	t.Run("up on Title stays on Title", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "kept", Description: "", Content: ""}).
			Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), special(tea.KeyUp))
		screen.press(typed("kept")...)
		screen.press(ctrl('s'))
	})

	t.Run("saves from inside Content", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "", Description: "", Content: "body"}).
			Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(enterContent()...)
		screen.press(typed("body")...)
		screen.press(ctrl('s'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func TestModel_editOverlayPendingSave(t *testing.T) {
	t.Parallel()

	t.Run("a second save while one is pending creates the Snippet once", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 1)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "x", Description: "", Content: ""}).
			Return(snippets[0], nil).
			Once()
		screen := start(t, savingModel(t, creator, snippets), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'))
		screen.hold()
		screen.press(ctrl('s'), ctrl('s'))
		screen.release()

		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("esc while a save is pending leaves no confirmation after it succeeds", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 1)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(snippets[0], nil)
		screen := start(t, savingModel(t, creator, snippets), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'))
		screen.hold()
		screen.press(ctrl('s'), special(tea.KeyEscape))
		screen.release()

		assert.NotContains(t, screen.screen(), discardQuestion)
		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("a rejected save can be saved again", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, mock.Anything).
			Return(domain.Snippet{}, domain.OnField(domain.FieldTitle, value.ErrBlankTitle)).
			Twice()
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('s'), ctrl('s'))
	})

	t.Run("a rejected save can be cancelled", func(t *testing.T) {
		t.Parallel()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, errDatabaseLocked)
		screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('s'), special(tea.KeyEscape))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})
}

func TestModel_editOverlaySaveFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status string
	}{
		{
			name:   "names a blank Title",
			err:    domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
			status: "Title is blank",
		},
		{
			name:   "names a long Title",
			err:    domain.OnField(domain.FieldTitle, value.ErrTitleTooLong),
			status: "Title is longer than 200 characters",
		},
		{
			name:   "names a long Description",
			err:    domain.OnField(domain.FieldDescription, value.ErrDescriptionTooLong),
			status: "Description is longer than 2000 characters",
		},
		{
			name:   "names large Content",
			err:    domain.OnField(domain.FieldContent, value.ErrContentTooLong),
			status: "Content is larger than 256 KiB",
		},
		{
			name: "names the first of several fields",
			err: errors.Join(
				domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
				domain.OnField(domain.FieldContent, value.ErrContentTooLong),
			),
			status: "Title is blank",
		},
		{
			name:   "shows a generic failure for a rule it has no words for",
			err:    domain.OnField(domain.FieldLanguage, value.ErrUnknownLanguage),
			status: "Something went wrong; see the log",
		},
		{
			name:   "shows a generic failure otherwise",
			err:    errDatabaseLocked,
			status: "Something went wrong; see the log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			creator := NewMockSnippetCreator(t)
			creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, tt.err)
			screen := start(t, editingModel(t, creator), wideWidth, wideHeight)

			screen.press(letter('n'), ctrl('s'))

			assert.Contains(t, screen.screen(), tt.status)
			assert.Contains(t, screen.screen(), editOverlayTitle)
		})
	}
}

func TestModel_editOverlayCancel(t *testing.T) {
	t.Parallel()

	esc := special(tea.KeyEscape)
	tests := []struct {
		name        string
		keys        []tea.KeyPressMsg
		wantOverlay bool
		wantText    string
	}{
		{name: "closes at once without changes", keys: []tea.KeyPressMsg{letter('n'), esc}},
		{
			name:        "asks before discarding changes",
			keys:        []tea.KeyPressMsg{letter('n'), letter('x'), esc},
			wantOverlay: true,
			wantText:    discardQuestion,
		},
		{name: "discards on y", keys: []tea.KeyPressMsg{letter('n'), letter('x'), esc, letter('y')}},
		{
			name:        "keeps the changes on n",
			keys:        []tea.KeyPressMsg{letter('n'), letter('x'), esc, letter('n')},
			wantOverlay: true,
			wantText:    unsavedTitle,
		},
		{
			name:        "esc inside Content leaves it and keeps the overlay",
			keys:        append(enterContent(), esc),
			wantOverlay: true,
			wantText:    contentEntryHint,
		},
		{name: "a second esc cancels", keys: append(enterContent(), esc, esc)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

			screen.press(tt.keys...)

			assert.Equal(t, tt.wantOverlay, strings.Contains(screen.screen(), editOverlayTitle))
			assert.Contains(t, screen.screen(), tt.wantText)
		})
	}
}

func TestModel_editOverlayForcedQuit(t *testing.T) {
	t.Parallel()

	t.Run("asks before quitting with changes", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'), ctrl('c'))

		assert.Contains(t, screen.screen(), quitQuestion)
		assert.Contains(t, screen.screen(), "y yes · n no")
		assert.NotContains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("quits on y", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'), ctrl('c'), letter('y'))

		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("quits at once without changes", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), ctrl('c'))

		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("ignores other keys and pastes while asking", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'), letter('x'), ctrl('c'), letter('z'))
		screen.send(tea.PasteMsg{Content: "pasted"})

		assert.Contains(t, screen.screen(), quitQuestion)
		assert.NotContains(t, screen.screen(), "pasted")
	})
}

func TestModel_editOverlayPaste(t *testing.T) {
	t.Parallel()

	t.Run("refuses a paste with tabs into Content", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(enterContent()...)
		screen.send(tea.PasteMsg{Content: "if x {\n\treturn\n}"})

		assert.Contains(t, screen.screen(), "Pasted text contains tabs; use ctrl+e to edit in $EDITOR")
		assert.NotContains(t, screen.screen(), "return")
		assert.NotContains(t, screen.screen(), unsavedTitle)
	})

	t.Run("inserts a paste without tabs into Content", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(enterContent()...)
		screen.send(tea.PasteMsg{Content: "if x {\n    return\n}"})

		assert.Contains(t, screen.screen(), "return")
	})

	t.Run("pastes into Title", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.press(letter('n'))
		screen.send(tea.PasteMsg{Content: "Pasted title"})

		assert.Contains(t, screen.screen(), "Pasted title")
	})

	t.Run("ignores a paste on the main screen", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		screen.send(tea.PasteMsg{Content: "Pasted title"})

		assert.NotContains(t, screen.screen(), "Pasted title")
	})
}

func editingModel(t *testing.T, creator *MockSnippetCreator) tui.Model {
	t.Helper()

	return creatingModel(t, creator, listerOf(t))
}

func savingModel(t *testing.T, creator *MockSnippetCreator, saved []domain.Snippet) tui.Model {
	t.Helper()

	return creatingModel(t, creator, listerReturning(t, nil, saved))
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

func enterContent() []tea.KeyPressMsg {
	return []tea.KeyPressMsg{letter('n'), special(tea.KeyDown), special(tea.KeyDown), special(tea.KeyEnter)}
}
