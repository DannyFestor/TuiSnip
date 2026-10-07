package folderpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_UpdateCollapse(t *testing.T) {
	t.Parallel()

	t.Run("space collapses the Folder under the cursor and hides its subtree", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('j'), keypress.Letter('j'))

		pane, outcomes, _ := pane.Update(keypress.Letter(' '))

		assert.Equal(t, []string{
			"◆ Root                2",
			"  docker              3",
			"▸ GO                  2",
		}, viewLines(pane))
		assert.Equal(t, []outcome.Outcome{
			outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{sample.Go.ID()}},
		}, outcomes)
	})

	t.Run("space again expands it", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('j'), keypress.Letter('j'), keypress.Letter(' '))

		pane, outcomes, _ := pane.Update(keypress.Letter(' '))

		assert.Equal(t, []string{
			"◆ Root                2",
			"  docker              3",
			"▾ GO                  2",
			"    testing           1",
		}, viewLines(pane))
		assert.Equal(t, []outcome.Outcome{outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{}}}, outcomes)
	})

	t.Run("leaves the Root and a Folder without subfolders alone", func(t *testing.T) {
		t.Parallel()

		pane := samplePane(t, foldertree.New(t))

		pane, atRoot, _ := pane.Update(keypress.Letter(' '))
		pane, onLeaf, _ := pressed(pane, keypress.Letter('j')).Update(keypress.Letter(' '))

		assert.Empty(t, atRoot)
		assert.Empty(t, onLeaf)
		assert.Len(t, viewLines(pane), 4)
	})

	t.Run("moves the cursor over the visible rows only", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := collapsedPane(t, sample, sample.Go.ID())

		pane = pressed(pane, keypress.Letter('G'))

		assert.Equal(t, sample.Go.ID(), pane.Selected())
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("starts with the remembered Folders collapsed", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)

		pane := collapsedPane(t, sample, sample.Go.ID())

		assert.Equal(t, []string{
			"◆ ROOT                2",
			"  docker              3",
			"▸ go                  2",
		}, viewLines(pane))
	})
}

func TestPane_WithTreeCollapsed(t *testing.T) {
	t.Parallel()

	t.Run("drops a collapsed id no Folder has once the tree arrives", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		stale := testkit.NewSequentialIDs().NewFolderID()
		pane := treelessPane(t, stale, sample.Go.ID())

		pane, outcomes := pane.WithTree(sample.Tree)

		assert.Equal(t, []outcome.Outcome{
			outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{sample.Go.ID()}},
		}, outcomes)
		assert.Contains(t, viewLines(pane), "▸ go                  2")
	})

	t.Run("reports nothing when every collapsed id has its Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := treelessPane(t, sample.Go.ID())

		_, outcomes := pane.WithTree(sample.Tree)

		assert.Empty(t, outcomes)
	})
}

func TestPane_WithCursorOnCollapsed(t *testing.T) {
	t.Parallel()

	t.Run("expands the collapsed Folders that hide the Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := collapsedPane(t, sample, sample.Go.ID())

		pane, outcomes := pane.WithCursorOn(sample.Tests.ID())

		assert.Equal(t, sample.Tests.ID(), pane.Selected())
		assert.Equal(t, []outcome.Outcome{outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{}}}, outcomes)
		assert.Contains(t, viewLines(pane), "▾ go                  2")
	})

	t.Run("leaves a visible Folder's collapsed state alone", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := collapsedPane(t, sample, sample.Go.ID())

		pane, outcomes := pane.WithCursorOn(sample.Go.ID())

		assert.Equal(t, sample.Go.ID(), pane.Selected())
		assert.Empty(t, outcomes)
	})
}

func TestPane_UpdateNewFolderInCollapsed(t *testing.T) {
	t.Parallel()

	sample := foldertree.New(t)
	pane := pressed(collapsedPane(t, sample, sample.Go.ID()), keypress.Letter('G'))

	pane, outcomes, _ := pane.Update(keypress.Letter('N'))

	assert.Equal(t, []outcome.Outcome{outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{}}}, outcomes)
	assert.Contains(t, viewLines(pane), "    testing           1")
}

func collapsedPane(t *testing.T, sample foldertree.Sample, collapsed ...domain.FolderID) folderpane.Pane {
	t.Helper()

	return withTree(treelessPane(t, collapsed...), sample.Tree)
}

func treelessPane(t *testing.T, collapsed ...domain.FolderID) folderpane.Pane {
	t.Helper()

	pane := folderpane.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), collapsed)
	pane, _, _ = pane.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: boxHeight}})

	return pane
}
