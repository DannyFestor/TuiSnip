package tui_test

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestModel_moveFolder(t *testing.T) {
	t.Parallel()

	t.Run("keeps the moved Folder as the Browse selection, expanding its new parent", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		saver := NewMockCollapsedFoldersSaver(t)
		saver.EXPECT().SaveCollapsedFolders(mock.Anything, []domain.FolderID{}).Return(nil).Once()

		lister := listerThrough(t, sample)
		listingIn(lister, sample.Go.ID())

		with := movingGoUnderDocker(t, sample, lister)
		with.collapsedFoldersSaver = saver
		settings := testsettings.Default(t)
		settings.Remembered.CollapsedFolders = []domain.FolderID{sample.Docker.ID()}
		screen := start(t, modelWithSettings(t, with, settings), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'), keypress.Letter('j'))
		pickedFolder(screen, "docker")

		assert.Contains(t, screen.screen(), "3 Root / docker / go · by title")
		assert.Regexp(t, `▾ docker`, screen.screen())
	})

	t.Run("takes the Browse selection back from a Tag", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		tags := sampleTagCounts(t)
		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle)

		lister := listerThrough(t, sample)
		listingIn(lister, sample.Go.ID())

		with := movingGoUnderDocker(t, sample, lister)
		with.tagLister = tagsOf(t, tags...)
		with.tagSnippetsLister = tagged
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'), keypress.Letter('j'))
		screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter), keypress.Letter('1'))
		pickedFolder(screen, "docker")

		assert.Contains(t, screen.screen(), "3 Root / docker / go · by title")
	})

	t.Run("reports a refused move with the generic failure", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		mover := NewMockFolderMover(t)
		mover.EXPECT().Run(mock.Anything, mock.Anything).
			Return(domain.Folder{}, fmt.Errorf("folder.Move: %w", domain.ErrFolderCycle))

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())

		with := browsingActions(t, sample, lister)
		with.folderMover = mover
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'))
		pickedFolder(screen, "go")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.Contains(t, screen.screen(), "3 Root / docker · by title")
	})
}

func TestModel_moveSnippet(t *testing.T) {
	t.Parallel()

	t.Run("in a Folder, the Snippet leaves the list and the one below takes its row", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		snippets := numberedSnippets(t, 3)
		mover := NewMockSnippetMover(t)
		mover.EXPECT().Run(mock.Anything, snippet.MoveInput{SnippetID: snippets[1].ID(), FolderID: sample.Docker.ID()}).
			Return(snippets[1].MoveTo(sample.Docker.ID()), nil)

		with := browsingActions(t, sample, listerReturning(t, snippets, []domain.Snippet{snippets[0], snippets[2]}))
		with.mover = mover
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'), keypress.Letter('j'))
		pickedFolder(screen, "docker")

		assert.NotContains(t, screen.screen(), "Snippet 2")
		assert.Contains(t, screen.screen(), "Description 3")
	})

	t.Run(
		"under a Tag, the Snippet stays listed under the cursor with its new Folder and its Language",
		func(t *testing.T) {
			t.Parallel()

			sample := foldertree.New(t)
			tags := sampleTagCounts(t)
			snippets := sampleSnippets(t)
			moved := snippets[1].MoveTo(sample.Docker.ID())
			mover := NewMockSnippetMover(t)
			mover.EXPECT().Run(mock.Anything, snippet.MoveInput{SnippetID: moved.ID(), FolderID: sample.Docker.ID()}).
				Return(moved, nil)

			tagged := NewMockTagSnippetsLister(t)
			listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, snippets...)
			listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, snippets[0], moved)

			with := browsingActions(t, sample, listerOf(t))
			with.tagLister = tagsOf(t, tags...)
			with.tagSnippetsLister = tagged
			with.mover = mover
			screen := start(t, modelWith(t, with), wideWidth, wideHeight)

			screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter), keypress.Letter('j'))
			pickedFolder(screen, "docker")

			assert.Contains(t, screen.screen(), "3 # docker · by title")
			assert.Contains(t, screen.screen(), "Reclaim disk space")
			assert.Contains(t, screen.screen(), "Root / docker · Bash")
		},
	)

	t.Run("reports a failed move", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		mover := NewMockSnippetMover(t)
		mover.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Snippet{}, errDatabaseLocked)

		with := browsingActions(t, sample, listerOf(t, sampleSnippets(t)...))
		with.mover = mover
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('3'))
		pickedFolder(screen, "docker")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func pickedFolder(screen *driver, filter string) {
	screen.t.Helper()

	screen.press(keypress.Letter('m'))
	screen.press(keypress.Typed(filter)...)
	screen.press(keypress.Special(tea.KeyEnter))
}

func movingGoUnderDocker(t *testing.T, sample foldertree.Sample, lister *MockFolderSnippetsLister) actions {
	t.Helper()

	moved, err := sample.Go.MoveUnder(sample.Docker.ID(), nil)
	require.NoError(t, err)

	mover := NewMockFolderMover(t)
	mover.EXPECT().Run(mock.Anything, folder.MoveInput{FolderID: sample.Go.ID(), ParentID: sample.Docker.ID()}).
		Return(moved, nil)

	return actions{
		lister:      lister,
		treeLister:  treesOf(t, sample.Tree, goUnderDocker(sample, moved)),
		copier:      NewMockSnippetCopier(t),
		creator:     NewMockSnippetCreator(t),
		searcher:    NewMockSnippetSearcher(t),
		folderMover: mover,
	}
}

func goUnderDocker(sample foldertree.Sample, moved domain.Folder) browse.Tree {
	tree := sample.Tree
	goNode := tree.Folders[1]
	goNode.Folder = moved
	dockerNode := tree.Folders[0]
	dockerNode.Children = []browse.FolderNode{goNode}
	tree.Folders = []browse.FolderNode{dockerNode}

	return tree
}
