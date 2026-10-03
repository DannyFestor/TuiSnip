package editoverlay_test

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	overlayTitle     = "Editing"
	unsavedTitle     = "Editing •"
	contentEntryHint = "enter or down to edit"
	discardQuestion  = "Discard the unsaved changes? y/N"
	quitQuestion     = "Quit and discard the unsaved changes? y/N"
	titleInputWidth  = 91
	contentRows      = 31
)

var errDatabaseLocked = errors.New("database is locked")

func screenSize() look.Size {
	return look.Size{Width: 120, Height: 40}
}

func editing(t *testing.T) *overlaytest.Driver {
	t.Helper()

	return editingWith(t, testsettings.Default(t).Keys)
}

func editingWith(t *testing.T, keys binding.Keys) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(keys, look.NewStyles())

	return overlaytest.Open(t, screenSize(), opened)
}

func enterContent() []tea.KeyPressMsg {
	return []tea.KeyPressMsg{
		overlaytest.Special(tea.KeyDown),
		overlaytest.Special(tea.KeyDown),
		overlaytest.Special(tea.KeyEnter),
	}
}

func save() tea.KeyPressMsg {
	return overlaytest.Ctrl('s')
}

func esc() tea.KeyPressMsg {
	return overlaytest.Special(tea.KeyEscape)
}

func savedSnippet(t *testing.T) domain.Snippet {
	t.Helper()

	return testkit.Snippet(t, testkit.SnippetSpec{ID: testkit.NewSequentialIDs().NewSnippetID()})
}

func input(title, description, content string) snippet.CreateInput {
	return snippet.CreateInput{Title: title, Description: description, Content: content}
}
