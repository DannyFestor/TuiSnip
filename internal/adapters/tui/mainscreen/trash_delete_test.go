package mainscreen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const (
	deletePermanentlyTitle = "Delete permanently"
	emptyTrashTitle        = "Empty Trash"
)

func TestScreen_trashDeletePermanently(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		trashItem func(t *testing.T) trash.Item
		want      string
	}{
		{
			name:      "d asks before deleting a trashed Snippet for good",
			trashItem: trashedSnippetItem,
			want:      `Permanently delete Snippet "Table test skeleton"? This cannot be undone. [y/N]`,
		},
		{
			name:      "d on a trashed Folder counts its subfolders and Snippets",
			trashItem: trashedFolderItem,
			want:      `Permanently delete Folder "docker", 1 subfolder, and 15 Snippets? This cannot be undone. [y/N]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			driver, _ := inTrash(t, tt.trashItem(t))
			driver.Press(keypress.Letter('3'))

			driver.Press(keypress.Letter('d'))

			assert.Contains(t, driver.Screen(), deletePermanentlyTitle)
			assert.Contains(t, driver.Screen(), tt.want)
			assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		})
	}

	t.Run("yes asks to delete the Trash item under the cursor", func(t *testing.T) {
		t.Parallel()

		trashItem := trashedFolderItem(t)
		driver, _ := inTrash(t, trashedSnippetItem(t), trashItem)
		driver.Press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('d'))

		driver.Press(keypress.Letter('y'))

		want := outcome.TrashDeleteRequested{
			Input: trash.DeletePermanentlyInput{Item: trashItem.ID()},
			Name:  trashedFolder,
		}
		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}, want}, driver.Outcomes())
	})

	t.Run("no keeps the Trash item", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t, trashedSnippetItem(t))
		driver.Press(keypress.Letter('3'), keypress.Letter('d'))

		driver.Press(keypress.Letter('n'))

		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		assert.NotContains(t, driver.Screen(), deletePermanentlyTitle)
	})

	t.Run("d in an empty Trash asks nothing", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t)
		driver.Press(keypress.Letter('3'))

		driver.Press(keypress.Letter('d'))

		assert.NotContains(t, driver.Screen(), deletePermanentlyTitle)
	})
}

func TestScreen_trashEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		trashItems func(t *testing.T) []trash.Item
		want       string
	}{
		{
			name: "D counts every Folder and Snippet, those inside trashed Folders included",
			trashItems: func(t *testing.T) []trash.Item {
				t.Helper()

				return []trash.Item{trashedSnippetItem(t), trashedFolderItem(t)}
			},
			want: "Permanently delete 2 Folders and 16 Snippets in the Trash? This cannot be undone. [y/N]",
		},
		{
			name: "D leaves out a zero count of Folders and says 1 Snippet for one",
			trashItems: func(t *testing.T) []trash.Item {
				t.Helper()

				return []trash.Item{trashedSnippetItem(t)}
			},
			want: "Permanently delete 1 Snippet in the Trash? This cannot be undone. [y/N]",
		},
		{
			name: "D leaves out a zero count of Snippets and says 1 Folder for one",
			trashItems: func(t *testing.T) []trash.Item {
				t.Helper()

				empty := trashedFolderItem(t)
				empty.Contents = trash.Contents{SubfolderCount: 0, SnippetCount: 0, Outline: trash.Outline{}}

				return []trash.Item{empty}
			},
			want: "Permanently delete 1 Folder in the Trash? This cannot be undone. [y/N]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			driver, _ := inTrash(t, tt.trashItems(t)...)
			driver.Press(keypress.Letter('3'))

			driver.Press(keypress.Letter('D'))

			assert.Contains(t, driver.Screen(), emptyTrashTitle)
			assert.Contains(t, driver.Screen(), tt.want)
			assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, driver.Outcomes())
		})
	}

	t.Run("yes asks to empty the Trash items the prompt counted", func(t *testing.T) {
		t.Parallel()

		snippetItem, folderItem := trashedSnippetItem(t), trashedFolderItem(t)
		driver, _ := inTrash(t, snippetItem, folderItem)
		driver.Press(keypress.Letter('3'), keypress.Letter('D'))

		driver.Press(keypress.Letter('y'))

		want := outcome.TrashEmptyRequested{
			Input: trash.EmptyInput{Items: []trash.ItemID{snippetItem.ID(), folderItem.ID()}},
		}
		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}, want}, driver.Outcomes())
	})

	t.Run("D in an empty Trash asks nothing", func(t *testing.T) {
		t.Parallel()

		driver, _ := inTrash(t)
		driver.Press(keypress.Letter('3'))

		driver.Press(keypress.Letter('D'))

		assert.NotContains(t, driver.Screen(), emptyTrashTitle)
	})
}
