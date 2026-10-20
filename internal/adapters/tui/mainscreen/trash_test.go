package mainscreen_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
)

const (
	trashListTitle  = "3 Trash"
	trashListHint   = "y Copy · r restore · z zoom · / search · ? help"
	deletedOnText   = "deleted 2026-10-14"
	trashedPath     = "Root / go / testing"
	trashedSnippet  = "Table test skeleton"
	trashedFolder   = "docker"
	trashedFolderOf = 15
)

func deletedAt() time.Time {
	return time.Date(2026, time.October, 14, 12, 0, 0, 0, time.UTC)
}

func trashedSnippetItem(t *testing.T) trash.Item {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return trash.Item{
		Kind: trash.ItemKindSnippet,
		Snippet: testkit.Snippet(t, testkit.SnippetSpec{
			ID:        ids.NewSnippetID(),
			Title:     trashedSnippet,
			Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID(), Language: "Go", Content: "package x\n"}},
		}),
		Folder:    domain.Folder{},
		DeletedAt: deletedAt(),
		Path:      []string{"go", "testing"},
		Contents:  trash.Contents{SubfolderCount: 0, SnippetCount: 0, Outline: trash.Outline{}},
	}
}

func trashedFolderItem(t *testing.T) trash.Item {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return trash.Item{
		Kind:    trash.ItemKindFolder,
		Snippet: domain.Snippet{},
		Folder: testkit.Folder(t, testkit.FolderSpec{
			ID:   ids.NewFolderID(),
			Name: trashedFolder,
		}),
		DeletedAt: deletedAt(),
		Path:      []string{"ops"},
		Contents: trash.Contents{
			SubfolderCount: 1,
			SnippetCount:   trashedFolderOf,
			Outline: trash.Outline{
				Name:          trashedFolder,
				SnippetTitles: []string{"Prune everything"},
				Subfolders:    []trash.Outline{{Name: "compose", SnippetTitles: nil, Subfolders: nil}},
			},
		},
	}
}

func inTrash(t *testing.T, items ...trash.Item) (*overlaytest.Driver, foldertree.Sample) {
	t.Helper()

	driver, sample := browsing(t)
	driver.Press(keypress.Letter('G'))
	driver.Send(mainscreen.TrashLoaded{Items: items})

	return driver, sample
}

func TestScreen_trashSelection(t *testing.T) {
	t.Parallel()

	t.Run("G in Folders lands on the Trash and reports it as the Browse selection", func(t *testing.T) {
		t.Parallel()

		driver, _ := browsing(t)

		driver.Press(keypress.Letter('G'))

		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		assert.Contains(t, driver.Screen(), "▪ Trash")
		assert.Contains(t, driver.Screen(), trashListTitle)
		assert.NotContains(t, driver.Screen(), "by title")
	})

	t.Run("leaving the Trash with k selects the last Folder", func(t *testing.T) {
		t.Parallel()

		driver, sample := browsing(t)
		driver.Press(keypress.Letter('G'))

		driver.Press(keypress.Letter('k'))

		assert.Equal(t, outcome.FolderSelected{ID: sample.Tests.ID()}, driver.Outcomes()[len(driver.Outcomes())-1])
	})

	t.Run("ignores a loaded Trash while the Trash is not the Browse selection", func(t *testing.T) {
		t.Parallel()

		driver, _ := browsing(t)

		driver.Send(mainscreen.TrashLoaded{Items: []trash.Item{trashedFolderItem(t)}})
		driver.Press(keypress.Letter('G'))

		assert.Contains(t, driver.Screen(), "The Trash is empty.")
		assert.NotContains(t, driver.Screen(), "▸ "+trashedFolder)
	})

	t.Run("saving a new Snippet from the Trash returns the Browse selection to the Root", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))
		saved := outcome.SnippetSaved{
			ID:        domain.SnippetID{},
			Selection: browseselection.InFolder(domain.FolderID{}),
		}

		driver.Offer(saved)

		assert.Contains(t, driver.Screen(), snippetListTitle)
		assert.Equal(t, saved, driver.Outcomes()[len(driver.Outcomes())-1])
	})

	t.Run("opens Search with a note that it does not cover the Trash", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))

		driver.Press(keypress.Letter('/'))

		assert.Contains(t, driver.Screen(), "Search doesn't cover the Trash")
	})

	t.Run("n opens the edit overlay", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))

		driver.Press(keypress.Letter('n'))

		assert.Contains(t, driver.Screen(), "Editing")
		assert.Equal(t, editorHint, driver.Hints())
	})
}

func TestScreen_trashListing(t *testing.T) {
	t.Parallel()

	t.Run("shows a trashed Folder with its Snippet count", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedFolderItem(t))

		assert.Regexp(t, `▸ docker +15 Snippets`, driver.Screen())
	})

	t.Run("shows an empty Trash in the list and in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t)

		assert.Equal(t, 2, strings.Count(driver.Screen(), "The Trash is empty."))
	})

	t.Run("shows a trashed Snippet with its path and deletion date in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))

		driver.Press(keypress.Typed("4z")...)

		assert.Contains(t, driver.Screen(), trashedPath+" · Go")
		assert.Contains(t, driver.Screen(), deletedOnText)
	})

	t.Run("shows a trashed Folder with its path, contents, deletion date and outline", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedFolderItem(t))

		assert.Contains(t, driver.Screen(), "Root / ops")
		assert.Contains(t, driver.Screen(), "1 subfolder · 15 Snippets")
		assert.Contains(t, driver.Screen(), deletedOnText)
		assert.Contains(t, driver.Screen(), "compose")
		assert.Contains(t, driver.Screen(), "Prune everything")
	})

	t.Run("hints only the Actions the Trash offers in the list", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))

		driver.Press(keypress.Letter('3'))

		assert.Equal(t, trashListHint, driver.Hints())
	})
}

func TestScreen_trashActions(t *testing.T) {
	t.Parallel()

	t.Run("y on a Snippet item asks to copy it from the Trash", func(t *testing.T) {
		t.Parallel()

		item := trashedSnippetItem(t)
		driver, _ := inTrash(t, item)
		driver.Press(keypress.Letter('3'))

		driver.Press(keypress.Letter('y'))

		want := outcome.CopyRequested{
			ID:         item.Snippet.ID(),
			FragmentID: item.Snippet.Fragments()[0].ID(),
			Title:      item.Snippet.Title().String(),
			InTrash:    true,
		}
		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}, want}, driver.Outcomes())
	})

	t.Run("y on a Folder item reports nothing and explains why", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedFolderItem(t))
		driver.Press(keypress.Letter('3'))

		driver.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		assert.Contains(t, statusLine(driver), "Only a Snippet can be copied")
	})

	t.Run("r asks to restore the selected item", func(t *testing.T) {
		t.Parallel()

		item := trashedSnippetItem(t)
		driver, _ := inTrash(t, item)
		driver.Press(keypress.Letter('3'))

		driver.Press(keypress.Letter('r'))

		want := outcome.TrashRestoreRequested{
			Input: trash.RestoreInput{Item: item.ID()},
			Name:  item.Name(),
			Path:  item.Path,
		}
		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}, want}, driver.Outcomes())
	})

	t.Run("the Snippet list keys that change Snippets report nothing", func(t *testing.T) {
		t.Parallel()

		for _, letter := range "eEmcds" {
			t.Run(string(letter), func(t *testing.T) {
				t.Parallel()

				driver, _ := inTrash(t, trashedSnippetItem(t))
				driver.Press(keypress.Letter('3'))

				driver.Press(keypress.Letter(letter))

				assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
			})
		}
	})

	t.Run("e and E in the Snippet pane say to restore the Snippet first", func(t *testing.T) {
		t.Parallel()

		for _, letter := range "eE" {
			t.Run(string(letter), func(t *testing.T) {
				t.Parallel()

				driver, _ := inTrash(t, trashedSnippetItem(t))
				driver.Press(keypress.Letter('4'))

				driver.Press(keypress.Letter(letter))

				assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
				assert.Contains(t, statusLine(driver), "Restore the Snippet to edit it")
			})
		}
	})

	t.Run("y in the Snippet pane on a Folder item says only a Snippet can be copied", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedFolderItem(t))
		driver.Press(keypress.Letter('4'))

		driver.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		assert.Contains(t, statusLine(driver), "Only a Snippet can be copied")
	})
}

func TestScreen_goneElsewhere(t *testing.T) {
	t.Parallel()

	t.Run("shows the text and asks to reload the current Folder with the selected Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		driver := showing(t, wide(), snippets...)

		driver.Send(mainscreen.GoneElsewhereShown{Text: "Deleted elsewhere"})

		assert.Contains(t, statusLine(driver), "Deleted elsewhere")
		assert.Equal(t, []outcome.Outcome{
			outcome.ReloadAsked{Selection: browseselection.Selection{}, Selecting: snippets[0].ID()},
		}, driver.Outcomes())
	})

	t.Run("asks to reload the Trash with no Snippet to keep selected", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedFolderItem(t))

		driver.Send(mainscreen.GoneElsewhereShown{Text: "Restored elsewhere"})

		require.Contains(t, statusLine(driver), "Restored elsewhere")
		assert.Equal(t, outcome.ReloadAsked{Selection: browseselection.InTrash(), Selecting: domain.SnippetID{}},
			driver.Outcomes()[len(driver.Outcomes())-1])
	})
}
