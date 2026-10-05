package editoverlay_test

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
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

	opened, _ := editoverlay.New(keys, look.NewStyles())

	return overlaytest.Open(t, screenSize(), opened)
}

func editingStored(t *testing.T, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return editingStoredWith(t, testsettings.Default(t).Keys, stored)
}

func editingStoredWith(t *testing.T, keys binding.Keys, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Editing(
		keys,
		look.NewStyles(),
		editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed()},
		look.DarkCodeStyle,
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func browsed() browseselection.Selection {
	return browseselection.WithTag(testkit.NewSequentialIDs().NewTagID())
}

func storedSnippet(t *testing.T, content string) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:          ids.NewSnippetID(),
		Title:       "Prune",
		Description: "Reclaim",
		FolderID:    ids.NewFolderID(),
		Fragment:    testkit.FragmentSpec{Language: "Go", Content: content},
		UpdatedAt:   time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC),
	})
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
