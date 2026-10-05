//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	rootTitle   = "Prune at the Root"
	filedTitle  = "Table test skeleton"
	filedHeader = "Root / go / testing · plaintext"
)

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

func seedNestedFolders(t *testing.T, app *bootstrap.App) domain.Folder {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	golang := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go"})
	tests := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "testing", ParentID: golang.ID()})
	testapp.SeedFolder(t, app, golang)
	testapp.SeedFolder(t, app, tests)
	testapp.SeedSnippet(t, app, titledSnippet(t, ids, rootTitle, domain.FolderID{}))
	testapp.SeedSnippet(t, app, titledSnippet(t, ids, filedTitle, tests.ID()))

	return golang
}

func titledSnippet(t *testing.T, ids *testkit.SequentialIDs, title string, folderID domain.FolderID) domain.Snippet {
	t.Helper()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:       ids.NewSnippetID(),
		Title:    title,
		FolderID: folderID,
		Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID()},
	})
}
