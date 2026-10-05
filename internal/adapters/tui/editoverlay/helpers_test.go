package editoverlay_test

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
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
	contentRows      = 30
	pickerHints      = "down move · enter pick · esc close"

	languagePickerTitle = "Pick a Language"
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

	opened, _ := editoverlay.New(keys, styles, nil, destination())

	return overlaytest.Open(t, screenSize(), opened)
}

func editingIn(t *testing.T, filedIn editoverlay.Destination) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), nil, filedIn)

	return overlaytest.Open(t, screenSize(), opened)
}

func destinationIn(t *testing.T, languageName string) editoverlay.Destination {
	t.Helper()

	filedIn := destination()
	filedIn.Language = language(t, languageName)

	return filedIn
}

func editingOffering(t *testing.T, curated []value.Language) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(
		testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), curated, destination(),
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func editingStoredOffering(t *testing.T, curated []value.Language, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Editing(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		curated,
		editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed()},
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func capturing(t *testing.T, captured string) *overlaytest.Driver {
	t.Helper()

	return capturingIn(t, destination(), captured)
}

func capturingIn(t *testing.T, filedIn editoverlay.Destination, captured string) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Capturing(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		nil,
		editoverlay.Captured{Destination: filedIn, Content: captured},
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func destination() editoverlay.Destination {
	return editoverlay.Destination{
		Selection: browseselection.InFolder(destinationFolderID()),
		Tags:      nil,
		Language:  value.PlainText(),
	}
}

func destinationFolderID() domain.FolderID {
	return testkit.NewSequentialIDs().NewFolderID()
}

func editingStored(t *testing.T, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return editingStoredWith(t, testsettings.Default(t).Keys, stored)
}

func editingStoredWith(t *testing.T, keys binding.Keys, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return editingStoredStyled(t, keys, look.NewStyles(look.SchemeDark), stored)
}

func editingStoredStyled(
	t *testing.T,
	keys binding.Keys,
	styles look.Styles,
	stored domain.Snippet,
) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Editing(
		keys, styles, nil, editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed()},
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
	return append(toContent(), keypress.Special(tea.KeyEnter))
}

func toContent() []tea.KeyPressMsg {
	return append(toLanguage(), keypress.Special(tea.KeyDown))
}

func toLanguage() []tea.KeyPressMsg {
	return []tea.KeyPressMsg{keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown)}
}

func pickLanguage() tea.KeyPressMsg {
	return keypress.Ctrl('l')
}

func language(t *testing.T, name string) value.Language {
	t.Helper()

	parsed, err := value.NewLanguage(name)
	require.NoError(t, err)

	return parsed
}

func withoutLanguageRow(screen string) string {
	lines := strings.Split(screen, "\n")

	return strings.Join(slices.DeleteFunc(lines, func(line string) bool {
		return strings.Contains(line, "Language")
	}), "\n")
}

func shiftTab() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
}

func shiftUp() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModShift}
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
	return inputIn(title, description, value.PlainText().String(), content)
}

func inputIn(title, description, language, content string) snippet.CreateInput {
	return snippet.CreateInput{
		Title:       title,
		Description: description,
		Language:    language,
		Content:     content,
		FolderID:    destinationFolderID(),
		Tags:        nil,
	}
}
