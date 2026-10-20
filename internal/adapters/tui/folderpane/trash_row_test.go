package folderpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const trashedCount = 3

func TestPane_TrashRow(t *testing.T) {
	t.Parallel()

	t.Run("ends the pane with the Trash count", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		sample.Tree.TrashCount = trashedCount

		lines := viewLines(samplePane(t, sample))

		assert.Equal(t, "▪ Trash               3", lines[len(lines)-1])
	})

	t.Run("bottom lands on the Trash and reports it", func(t *testing.T) {
		t.Parallel()

		pane := samplePane(t, foldertree.New(t))

		pane, outcomes, _ := pane.Update(keypress.Letter('G'))

		assert.Equal(t, []outcome.Outcome{outcome.TrashSelected{}}, outcomes)
		assert.Equal(t, browseselection.InTrash(), pane.Selection())
	})

	t.Run("up from the Trash reports the last visible Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('G'))

		_, outcomes, _ := pane.Update(keypress.Letter('k'))

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Tests.ID()}}, outcomes)
	})

	t.Run("gives no Folder on the Trash", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('G'))

		_, ok := pane.SelectedFolder()

		assert.False(t, ok)
		assert.True(t, pane.Selected().IsNil())
	})

	for _, key := range []rune{'N', 'r', 'd', ' '} {
		t.Run("ignores "+string(key)+" on the Trash", func(t *testing.T) {
			t.Parallel()

			pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('G'))

			pane, outcomes, _ := pane.Update(keypress.Letter(key))

			assert.Empty(t, outcomes)
			assert.False(t, pane.Naming())
		})
	}

	t.Run("keeps the cursor on the Trash when the tree reloads", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), keypress.Letter('G'))
		reloaded := foldertree.New(t).Tree
		reloaded.TrashCount = trashedCount

		pane = withTree(pane, reloaded)

		assert.Equal(t, browseselection.InTrash(), pane.Selection())
	})
}
