//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const filedHeader = "Root / go / testing · plaintext"

func TestBrowseIntoNestedFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.press(keypress.Letter('G'))

	screen.waitForFrame(filedHeader)
	screen.waitForFrame("3 Root / go / testing")
}

func TestSearchRevealsSnippetInItsFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('/'))
	screen.press(keypress.Typed("skeleton")...)
	screen.waitForFrame("go / testing")
	screen.press(keypress.Special(tea.KeyEnter))

	screen.waitForFrame(filedHeader)
	screen.waitForFrame("3 Root / go / testing")
}

func TestBrowseByTagAcrossFolders(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	ids := testkit.NewSequentialIDs()
	golang := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	kubernetes := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "kubernetes"})
	testapp.SeedFolder(t, app, golang)
	testapp.SeedTag(t, app, kubernetes)
	carrying := []domain.Tag{kubernetes}
	testapp.SeedSnippet(t, app, snippetWithIDs(t, ids, testkit.SnippetSpec{Title: "Pods at the Root", Tags: carrying}))
	testapp.SeedSnippet(t, app, snippetWithIDs(t, ids, testkit.SnippetSpec{
		Title: "Client in go", FolderID: golang.ID(), Tags: carrying,
	}))
	screen := open(t, app)
	screen.waitForFrame("# kubernetes")

	screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

	screen.waitForFrame("3 # kubernetes · by title")
	screen.waitForFrame("Client in go")
	screen.waitForFrame("Root / go · plaintext · #kubernetes")
}

func TestCreateThenRenameFolder(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	seedNestedFolders(t, app)
	screen := open(t, app)
	screen.waitForFrame("    testing")

	screen.press(keypress.Letter('j'), keypress.Letter('N'))
	screen.press(keypress.Typed("errors")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("3 Root / go / errors")

	screen.press(keypress.Letter('r'))
	screen.press(keypress.Typed("-wrapping")...)
	screen.press(keypress.Special(tea.KeyEnter))

	screen.waitForFrame("3 Root / go / errors-wrapping")
}
