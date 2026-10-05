package folderpane_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the Root with no Snippets before the tree arrives", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

		assert.Equal(t, []string{"◆ ROOT                0"}, viewLines(pane))
	})

	t.Run("shows the Root above the indented tree with each Snippet count", func(t *testing.T) {
		t.Parallel()

		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), foldertree.New(t).Tree)

		assert.Equal(t, []string{
			"◆ ROOT                2",
			"  docker              3",
			"▾ go                  2",
			"    testing           1",
		}, viewLines(pane))
	})

	t.Run("cuts a row to a narrow box", func(t *testing.T) {
		t.Parallel()

		pane := withTree(
			paneIn(t, look.Size{Width: 6, Height: boxHeight}),
			browse.Tree{RootSnippetCount: 42, Folders: nil},
		)

		assert.Equal(t, []string{"◆ … 42"}, viewLines(pane))
	})

	t.Run("scrolls to keep the cursor in a short box", func(t *testing.T) {
		t.Parallel()

		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: 2}), foldertree.New(t).Tree)

		pane = pressed(pane, keypress.Letter('G'))

		assert.Equal(t, []string{"▾ go                  2", "    TESTING           1"}, viewLines(pane))
	})
}

func TestPane_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []tea.KeyPressMsg
		want func(foldertree.Sample) domain.FolderID
	}{
		{
			name: "down selects the next Folder",
			keys: []tea.KeyPressMsg{keypress.Letter('j')},
			want: func(sample foldertree.Sample) domain.FolderID { return sample.Docker.ID() },
		},
		{
			name: "bottom selects the last Folder",
			keys: []tea.KeyPressMsg{keypress.Letter('G')},
			want: func(sample foldertree.Sample) domain.FolderID { return sample.Tests.ID() },
		},
		{
			name: "top selects the Root",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Letter('g')},
			want: func(foldertree.Sample) domain.FolderID { return domain.FolderID{} },
		},
		{
			name: "page down moves a page",
			keys: []tea.KeyPressMsg{keypress.Special(tea.KeyPgDown)},
			want: func(sample foldertree.Sample) domain.FolderID { return sample.Tests.ID() },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sample := foldertree.New(t)
			pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)

			pane = pressed(pane, tt.keys...)

			assert.Equal(t, tt.want(sample), pane.Selected())
		})
	}

	t.Run("reports the Folder the cursor moves onto", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)

		_, outcomes, _ := pane.Update(keypress.Letter('j'))

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}}, outcomes)
	})

	t.Run("reports nothing when the cursor stays", func(t *testing.T) {
		t.Parallel()

		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), foldertree.New(t).Tree)

		_, outcomes, _ := pane.Update(keypress.Letter('k'))

		assert.Empty(t, outcomes)
	})

	t.Run("ignores keys that do not move", func(t *testing.T) {
		t.Parallel()

		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), foldertree.New(t).Tree)

		_, outcomes, _ := pane.Update(keypress.Letter('x'))

		assert.Empty(t, outcomes)
	})
}

func TestPane_WithCursorOn(t *testing.T) {
	t.Parallel()

	t.Run("selects the Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)

		pane = withCursorOn(pane, sample.Tests.ID())

		assert.Equal(t, sample.Tests.ID(), pane.Selected())
	})

	t.Run("selects the Root", func(t *testing.T) {
		t.Parallel()

		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), foldertree.New(t).Tree)
		pane = pressed(pane, keypress.Letter('G'))

		pane = withCursorOn(pane, domain.FolderID{})

		assert.Equal(t, domain.FolderID{}, pane.Selected())
	})

	t.Run("keeps the cursor for a Folder missing from the tree", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		sample := foldertree.New(t)
		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)
		pane = pressed(pane, keypress.Letter('j'))

		pane = withCursorOn(pane, ids.NewFolderID())

		assert.Equal(t, sample.Docker.ID(), pane.Selected())
	})
}

func TestPane_WithTree(t *testing.T) {
	t.Parallel()

	t.Run("keeps the selected Folder when the tree reloads", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		sample := foldertree.New(t)
		pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), sample.Tree)
		pane = withCursorOn(pane, sample.Go.ID())
		added := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "awk"})
		grown := sample.Tree
		grown.Folders = append([]browse.FolderNode{{Folder: added, SnippetCount: 0, Children: nil}}, grown.Folders...)

		pane = withTree(pane, grown)

		assert.Equal(t, sample.Go.ID(), pane.Selected())
	})
}

func TestPane_DefaultLanguageOf(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	parent := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go", DefaultLanguage: "Go"})
	child := testkit.Folder(t, testkit.FolderSpec{
		ID: ids.NewFolderID(), Name: "scripts", ParentID: parent.ID(), DefaultLanguage: "Bash",
	})
	tree := browse.Tree{RootSnippetCount: 0, Folders: []browse.FolderNode{{
		Folder:       parent,
		SnippetCount: 0,
		Children:     []browse.FolderNode{{Folder: child, SnippetCount: 0, Children: nil}},
	}}}
	pane := withTree(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), tree)

	tests := []struct {
		name   string
		folder domain.FolderID
		want   value.Language
	}{
		{name: "a top-level Folder's", folder: parent.ID(), want: parent.DefaultLanguage()},
		{name: "a nested Folder's", folder: child.ID(), want: child.DefaultLanguage()},
		{name: "plaintext for the Root", folder: domain.FolderID{}, want: value.PlainText()},
		{name: "plaintext for a Folder missing from the tree", folder: ids.NewFolderID(), want: value.PlainText()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, pane.DefaultLanguageOf(tt.folder))
		})
	}
}

func TestPane_UpdateDelete(t *testing.T) {
	t.Parallel()

	t.Run("asks to delete the Folder under the cursor", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := pressed(samplePane(t, sample), keypress.Letter('j'))

		_, outcomes, _ := pane.Update(keypress.Letter('d'))

		assert.Equal(t, []outcome.Outcome{outcome.FolderDeleteAsked{ID: sample.Docker.ID()}}, outcomes)
	})

	t.Run("leaves the Root alone", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := samplePane(t, foldertree.New(t)).Update(keypress.Letter('d'))

		assert.Empty(t, outcomes)
	})
}

func TestPane_ShortHelp(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeFolders).ShortHelp(), pane.ShortHelp())
	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeFolders).FullHelp(), pane.FullHelp())
}
