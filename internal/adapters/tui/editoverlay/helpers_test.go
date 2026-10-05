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
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	overlayTitle     = "Editing"
	unsavedTitle     = "Editing •"
	contentEntryHint = "enter or down to edit"
	discardQuestion  = "Discard the unsaved changes? [y/N]"
	quitQuestion     = "Quit and discard the unsaved changes? [y/N]"
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

	return editingStyled(t, keys, look.NewStyles(look.SchemeDark))
}

func editingStyled(t *testing.T, keys binding.Keys, styles look.Styles) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(keys, styles)

	return overlaytest.Open(t, screenSize(), opened)
}

func enterContent() []tea.KeyPressMsg {
	return []tea.KeyPressMsg{
		keypress.Special(tea.KeyDown),
		keypress.Special(tea.KeyDown),
		keypress.Special(tea.KeyEnter),
	}
}

func save() tea.KeyPressMsg {
	return keypress.Ctrl('s')
}

func savedSnippet(t *testing.T) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), FolderID: ids.NewFolderID()})
}

func input(title, description, content string) snippet.CreateInput {
	return snippet.CreateInput{Title: title, Description: description, Content: content}
}
