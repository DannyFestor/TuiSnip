package folderpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

func TestPane_UpdateClick(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the clicked row", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := samplePane(t, sample)

		pane, outcomes, _ := pane.Update(clickOn(t, pane, "docker"))

		assert.Equal(t, sample.Docker.ID(), pane.Selected())
		assert.Empty(t, outcomes)
	})

	t.Run("collapses the Folder whose marker is clicked and moves the cursor to it", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := samplePane(t, sample)

		pane, outcomes, _ := pane.Update(clickOn(t, pane, "▾"))

		assert.Equal(t, []string{
			"◆ Root                2",
			"  docker              3",
			"▸ GO                  2",
		}, viewLines(pane))
		assert.Equal(t, []outcome.Outcome{
			outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{sample.Go.ID()}},
		}, outcomes)
	})

	t.Run("expands a collapsed Folder whose marker is clicked", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := collapsedPane(t, sample, sample.Go.ID())

		pane, _, _ = pane.Update(clickOn(t, pane, "▸"))

		assert.Contains(t, viewLines(pane), "    testing           1")
	})

	t.Run("leaves the Folder alone on the second click of a double click", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := samplePane(t, sample)
		click := clickOn(t, pane, "▾")
		click.Double = true

		pane, outcomes, _ := pane.Update(click)

		assert.Equal(t, sample.Go.ID(), pane.Selected())
		assert.Empty(t, outcomes)
	})

	t.Run("does nothing below the last row", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('j'))
		below := clickOn(t, pane, "testing")
		below.At.Y++

		pane, outcomes, _ := pane.Update(below)

		assert.Equal(t, sample.Docker.ID(), pane.Selected())
		assert.Empty(t, outcomes)
	})

	t.Run("ignores a click while a name is typed", func(t *testing.T) {
		t.Parallel()

		pane := pressed(samplePane(t, foldertree.New(t)), newFolderTyped("ci")...)

		pane, _, _ = pane.Update(clickOn(t, pane, "docker"))

		assert.True(t, pane.Naming())
	})
}

func TestPane_HasRowAt(t *testing.T) {
	t.Parallel()

	pane := samplePane(t, foldertree.New(t))
	last := cellOf(t, pane, "testing")
	below := pointer.Point{X: last.X, Y: last.Y + 1}

	assert.True(t, pane.HasRowAt(last))
	assert.False(t, pane.HasRowAt(below))
}

func TestPane_UpdateWheel(t *testing.T) {
	t.Parallel()

	t.Run("scrolls the rows without moving the cursor", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: 2}), sample.Tree)

		pane, _, _ = pane.Update(wheelDown(t, pane))

		assert.Equal(t, []string{"▾ go                  2", "    testing           1"}, viewLines(pane))
		assert.True(t, pane.Selected().IsNil())
	})

	t.Run("keeps the scroll when resized to the same box", func(t *testing.T) {
		t.Parallel()

		box := look.Size{Width: boxWidth, Height: 2}
		pane := withTree(paneIn(t, box), foldertree.New(t).Tree)

		pane, _, _ = pane.Update(wheelDown(t, pane))
		pane, _, _ = pane.Update(look.Resized{Box: box})

		assert.Equal(t, []string{"▾ go                  2", "    testing           1"}, viewLines(pane))
	})
}

func cellOf(t *testing.T, pane folderpane.Pane, text string) pointer.Point {
	t.Helper()

	return screencell.Find(t, pane.View(upperCursor()), text)
}

func clickOn(t *testing.T, pane folderpane.Pane, text string) pointer.Clicked {
	t.Helper()

	return pointer.Clicked{At: cellOf(t, pane, text), Double: false}
}

func wheelDown(t *testing.T, pane folderpane.Pane) pointer.Wheeled {
	t.Helper()

	return pointer.Wheeled{At: cellOf(t, pane, "ROOT"), Lines: 3}
}
