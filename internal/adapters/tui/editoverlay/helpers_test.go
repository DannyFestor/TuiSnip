package editoverlay_test

import (
	"errors"
	"fmt"
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
	titleInputWidth  = 88
	contentRows      = 30
	pickerHints      = "down move · enter pick · esc close"

	languagePickerTitle = "Pick a Language"

	maxContentBytes = 256 << 10
	pasteTooLarge   = "Paste would make Content larger than 256 KiB"

	goneElsewhereTitle = "Gone elsewhere"
	folderGoneQuestion = `Folder "docker" was moved to the Trash or deleted in another TuiSnip. ` +
		"Save the Snippet at the Root instead?"
)

var (
	errDatabaseLocked = errors.New("database is locked")
	errFolderGone     = fmt.Errorf("snippet.Create: %w", domain.ErrFolderGone)
)

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

	opened, _ := editoverlay.New(keys, styles, editoverlay.Options{}, destination())

	return overlaytest.Open(t, screenSize(), opened)
}

func editingIn(t *testing.T, filedIn editoverlay.Destination) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(
		testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), editoverlay.Options{}, filedIn,
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func destinationInGo(t *testing.T) editoverlay.Destination {
	t.Helper()

	filedIn := destination()
	filedIn.Language = language(t, "Go")

	return filedIn
}

func editingOffering(t *testing.T, curated []value.Language) *overlaytest.Driver {
	t.Helper()

	return editingWithOptions(t, editoverlay.Options{Languages: curated, Tags: nil}, destination())
}

func editingWithOptions(
	t *testing.T, options editoverlay.Options, filedIn editoverlay.Destination,
) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), options, filedIn)

	return overlaytest.Open(t, screenSize(), opened)
}

func editingStoredOffering(t *testing.T, curated []value.Language, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return editingStoredWithOptions(t, editoverlay.Options{Languages: curated, Tags: nil}, stored)
}

func editingStoredWithOptions(t *testing.T, options editoverlay.Options, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Editing(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		options,
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
		editoverlay.Options{},
		editoverlay.Captured{Destination: filedIn, Content: captured},
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func destination() editoverlay.Destination {
	return editoverlay.Destination{
		Selection:  browseselection.InFolder(destinationFolderID()),
		FolderName: "docker",
		Tags:       nil,
		Language:   value.PlainText(),
	}
}

func destinationFolderID() domain.FolderID {
	return testkit.NewSequentialIDs().NewFolderID()
}

func editingStored(tb testing.TB, stored domain.Snippet) *overlaytest.Driver {
	tb.Helper()

	return editingStoredWith(tb, testsettings.Default(tb).Keys, stored)
}

func editingStoredWith(tb testing.TB, keys binding.Keys, stored domain.Snippet) *overlaytest.Driver {
	tb.Helper()

	return editingStoredStyled(tb, keys, look.NewStyles(look.SchemeDark), stored)
}

func editingStoredStyled(
	tb testing.TB,
	keys binding.Keys,
	styles look.Styles,
	stored domain.Snippet,
) *overlaytest.Driver {
	tb.Helper()

	opened, _ := editoverlay.Editing(
		keys, styles, editoverlay.Options{}, editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed()},
	)

	return overlaytest.Open(tb, screenSize(), opened)
}

func browsed() browseselection.Selection {
	return browseselection.WithTag(testkit.NewSequentialIDs().NewTagID())
}

func storedSnippet(tb testing.TB, content string) domain.Snippet {
	tb.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(tb, testkit.SnippetSpec{
		ID:          ids.NewSnippetID(),
		Title:       "Prune",
		Description: "Reclaim",
		FolderID:    ids.NewFolderID(),
		Fragments:   []testkit.FragmentSpec{{Language: "Go", Content: content}},
		UpdatedAt:   time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC),
	})
}

func twoContentLines() []tea.KeyPressMsg {
	return append(enterContent(), keypress.Letter('a'), keypress.Special(tea.KeyEnter), keypress.Letter('b'))
}

func enterContent() []tea.KeyPressMsg {
	return append(toContent(), keypress.Special(tea.KeyEnter))
}

func toContent() []tea.KeyPressMsg {
	return append(toLanguage(), keypress.Special(tea.KeyDown))
}

func toLanguage() []tea.KeyPressMsg {
	return append(toFragmentTitle(), keypress.Special(tea.KeyDown))
}

func toFragmentTitle() []tea.KeyPressMsg {
	return append(toTags(), keypress.Special(tea.KeyDown))
}

func toTags() []tea.KeyPressMsg {
	return []tea.KeyPressMsg{keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown)}
}

func pickLanguage() tea.KeyPressMsg {
	return keypress.Ctrl('l')
}

func editTags() tea.KeyPressMsg {
	return keypress.Ctrl('t')
}

func language(t *testing.T, name string) value.Language {
	t.Helper()

	parsed, err := value.NewLanguage(name)
	require.NoError(t, err)

	return parsed
}

func twentyThousandShortLines() string {
	return strings.Join(slices.Repeat([]string{"line"}, 20_000), "\n")
}

func textOfSize(size int) string {
	const line = "\tfmt.Println(\"a line of typical code, indented by a tab\")\n"

	return strings.Repeat(line, size/len(line)) + line[:size%len(line)]
}

func shiftLeft() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModShift}
}

func shiftTab() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
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
		Fragments:   []snippet.FragmentInput{{Language: language, Content: content}},
		FolderID:    destinationFolderID(),
		Tags:        nil,
		NewTags:     nil,
	}
}
