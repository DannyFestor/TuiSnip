package mainscreen_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const pickerHints = "down move · enter pick · esc close"

func TestScreen_moveFolder(t *testing.T) {
	t.Parallel()

	t.Run("m on a Folder opens the Folder picker named for it", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('m'))

		assert.Contains(t, screen.Screen(), "Move go to")
		assert.Equal(t, pickerHints, screen.Hints())
	})

	t.Run("asks to move the Folder under the picked one", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('m'))
		screen.Press(keypress.Typed("docker")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.FolderMoveRequested{
			Input: folder.MoveInput{FolderID: sample.Go.ID(), ParentID: sample.Docker.ID()},
		}, lastOutcomeOf[outcome.FolderMoveRequested](t, screen.Outcomes()))
		assert.Equal(t, folderHints, screen.Hints())
	})

	t.Run("opens on the Folder's parent", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('m'))
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.FolderMoveRequested{
			Input: folder.MoveInput{FolderID: sample.Tests.ID(), ParentID: sample.Go.ID()},
		}, lastOutcomeOf[outcome.FolderMoveRequested](t, screen.Outcomes()))
	})

	t.Run("refuses to move the Folder into its own subtree", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('m'))
		screen.Press(keypress.Typed("go / testing")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.Screen(), "Move go to")
		assertNoMoveAsked(t, screen.Outcomes())
	})

	t.Run("m does nothing on the Root", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('m'))

		assert.NotContains(t, screen.Screen(), "Move ")
		assert.Equal(t, folderHints, screen.Hints())
	})

	t.Run("opens the picker with the configured key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeFolders][binding.Move] = []string{"ctrl+g"}
		screen := browsingWith(t, keys, nil)

		screen.Press(keypress.Letter('j'), keypress.Ctrl('g'))

		assert.Contains(t, screen.Screen(), "Move docker to")
	})
}

func TestScreen_moveSnippet(t *testing.T) {
	t.Parallel()

	t.Run("m in the Snippet list opens the Folder picker named for the Snippet", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('3'), keypress.Letter('m'))

		assert.Contains(t, screen.Screen(), "Move Graceful HTTP shutdown to")
		assert.Equal(t, pickerHints, screen.Hints())
	})

	t.Run("asks to move the Snippet into the picked Folder, the Snippet below taking its row", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		snippets := sampleSnippets(t)

		screen.Press(keypress.Letter('3'), keypress.Letter('m'))
		screen.Press(keypress.Typed("docker")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.SnippetMoveRequested{
			Input:     snippet.MoveInput{SnippetID: snippets[0].ID(), FolderID: sample.Docker.ID()},
			Selection: browseselection.InFolder(domain.FolderID{}),
			Selecting: snippets[1].ID(),
		}, lastOutcomeOf[outcome.SnippetMoveRequested](t, screen.Outcomes()))
	})

	t.Run("the Snippet above takes the row when the last row moves", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		snippets := sampleSnippets(t)

		screen.Press(keypress.Letter('3'), keypress.Letter('j'), keypress.Letter('m'))
		screen.Press(keypress.Typed("docker")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.SnippetMoveRequested{
			Input:     snippet.MoveInput{SnippetID: snippets[1].ID(), FolderID: sample.Docker.ID()},
			Selection: browseselection.InFolder(domain.FolderID{}),
			Selecting: snippets[0].ID(),
		}, lastOutcomeOf[outcome.SnippetMoveRequested](t, screen.Outcomes()))
	})

	t.Run("opens on the Snippet's Folder, and keeps the cursor on a Snippet that stays there", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)
		moving := sampleSnippets(t)[0]

		screen.Press(keypress.Letter('3'), keypress.Letter('m'), keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.SnippetMoveRequested{
			Input:     snippet.MoveInput{SnippetID: moving.ID(), FolderID: domain.FolderID{}},
			Selection: browseselection.InFolder(domain.FolderID{}),
			Selecting: moving.ID(),
		}, lastOutcomeOf[outcome.SnippetMoveRequested](t, screen.Outcomes()))
	})

	t.Run("keeps the cursor on the moved Snippet when a Tag is the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		sample := foldertree.New(t)
		moving := sampleSnippets(t)[0]

		screen.Send(mainscreen.TreeLoaded{Tree: sample.Tree})
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))
		screen.Send(mainscreen.SnippetsLoaded{
			Selection: browseselection.WithTag(tags[0].Tag.ID()),
			Snippets:  sampleSnippets(t),
			Selecting: domain.SnippetID{},
			Order:     domain.SortOrderTitle,
		})

		screen.Press(keypress.Letter('m'))
		screen.Press(keypress.Typed("docker")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, outcome.SnippetMoveRequested{
			Input:     snippet.MoveInput{SnippetID: moving.ID(), FolderID: sample.Docker.ID()},
			Selection: browseselection.WithTag(tags[0].Tag.ID()),
			Selecting: moving.ID(),
		}, lastOutcomeOf[outcome.SnippetMoveRequested](t, screen.Outcomes()))
	})

	t.Run("closing the picker asks for nothing", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('3'), keypress.Letter('m'), keypress.Special(tea.KeyEscape))

		assert.NotContains(t, screen.Screen(), "Move ")
		assertNoMoveAsked(t, screen.Outcomes())
	})

	t.Run("m does nothing in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('4'), keypress.Letter('m'))

		assert.NotContains(t, screen.Screen(), "Move ")
	})

	t.Run("m does nothing on an empty Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('3'), keypress.Letter('m'))

		assert.NotContains(t, screen.Screen(), "Move ")
	})
}

func assertNoMoveAsked(t *testing.T, outcomes []outcome.Outcome) {
	t.Helper()

	for _, reported := range outcomes {
		assert.IsNotType(t, outcome.FolderMoveRequested{}, reported)
		assert.IsNotType(t, outcome.SnippetMoveRequested{}, reported)
	}
}
