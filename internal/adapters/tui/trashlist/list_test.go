package trashlist_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/trashlist"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencell"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 30
	boxHeight = 3
)

func TestList_View(t *testing.T) {
	t.Parallel()

	t.Run("says the Trash is empty, with no Binding hints", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"The Trash is empty."}, rows(listOf(t)))
	})

	t.Run("shows a Snippet by title and Language and a Folder with its Snippet count", func(t *testing.T) {
		t.Parallel()

		got := rows(listOf(t, sampleTrashItems(t)...))

		assert.Equal(t, []string{
			"PRUNE EVERYTHING          BASH",
			"▸ docker           15 Snippets",
			"▸ empty              1 Snippet",
		}, got)
	})
}

func TestList_Update(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor with the movement keys", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, sampleTrashItems(t)...)

		list, _, _ = list.Update(keypress.Letter('G'))

		assert.Equal(t, "empty", selectedName(t, list))
	})

	t.Run("selects the clicked row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, sampleTrashItems(t)...)

		docker := screencell.Find(t, strings.Join(rows(list), "\n"), "docker")

		list, _, _ = list.Update(pointer.Clicked{At: docker, Double: false})

		assert.Equal(t, "docker", selectedName(t, list))
	})

	t.Run("keeps the cursor's row when the Trash items change", func(t *testing.T) {
		t.Parallel()

		trashItems := sampleTrashItems(t)
		list, _, _ := listOf(t, trashItems...).Update(keypress.Letter('j'))

		list = list.WithItems([]trash.Item{trashItems[0], trashItems[2]})

		assert.Equal(t, "empty", selectedName(t, list))
	})
}

func TestList_Selected(t *testing.T) {
	t.Parallel()

	_, ok := listOf(t).Selected()

	assert.False(t, ok)
}

func listOf(t *testing.T, trashItems ...trash.Item) trashlist.List {
	t.Helper()

	list := trashlist.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), languageOf)
	list, _, _ = list.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: boxHeight}})

	return list.WithItems(trashItems)
}

func languageOf(snippet domain.Snippet) string {
	return snippet.FirstFragment().Language().String()
}

func sampleTrashItems(t *testing.T) []trash.Item {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return []trash.Item{
		{
			Kind: trash.ItemKindSnippet,
			Snippet: testkit.Snippet(t, testkit.SnippetSpec{
				ID:        ids.NewSnippetID(),
				Title:     "Prune everything",
				Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID(), Language: "Bash"}},
			}),
		},
		{
			Kind:     trash.ItemKindFolder,
			Folder:   testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "docker"}),
			Contents: trash.Contents{SnippetCount: 15},
		},
		{
			Kind:     trash.ItemKindFolder,
			Folder:   testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "empty"}),
			Contents: trash.Contents{SnippetCount: 1},
		},
	}
}

func selectedName(t *testing.T, list trashlist.List) string {
	t.Helper()

	selected, ok := list.Selected()
	require.True(t, ok)

	return selected.Name()
}

func rows(list trashlist.List) []string {
	frame := look.FrameStyle{
		Border: lipgloss.NewStyle(),
		Title:  lipgloss.NewStyle(),
		Cursor: lipgloss.NewStyle().Transform(strings.ToUpper),
	}

	return strings.Split(ansi.Strip(list.View(frame)), "\n")
}
