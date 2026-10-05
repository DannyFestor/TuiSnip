package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_folders(t *testing.T) {
	t.Parallel()

	t.Run("creates a typed Folder and makes it the Browse selection", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		created := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "awk"})
		creator := NewMockFolderCreator(t)
		creator.EXPECT().Run(mock.Anything, folder.CreateInput{Name: "awk", ParentID: domain.FolderID{}}).
			Return(created, nil)

		lister := listerOf(t)
		listingIn(lister, created.ID(), filedIn(t, created.ID()))
		trees := treesOf(t, sample.Tree, withFolderAtRoot(sample.Tree, created))
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: trees, creator: creator, renamer: NewMockFolderRenamer(t),
		}), wideWidth, wideHeight)

		typedName(screen, 'N', "awk")

		assert.Contains(t, screen.screen(), "3 Root / awk · by title")
		assert.Contains(t, screen.screen(), filedTitle)
	})

	t.Run("renames the Folder under the cursor and shows the new name", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		renamed := testkit.Folder(t, testkit.FolderSpec{ID: sample.Docker.ID(), Name: "dockers"})
		renamer := NewMockFolderRenamer(t)
		renamer.EXPECT().Run(mock.Anything, folder.RenameInput{FolderID: sample.Docker.ID(), Name: "dockers"}).
			Return(renamed, nil)

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		reloaded := sample.Tree
		reloaded.Folders = append(
			[]browse.FolderNode{{Folder: renamed, SnippetCount: 0, Children: nil}},
			reloaded.Folders[1:]...)
		screen := start(t, folderModel(t, folderActions{
			lister:  lister,
			trees:   treesOf(t, sample.Tree, reloaded),
			creator: NewMockFolderCreator(t),
			renamer: renamer,
		}), wideWidth, wideHeight)
		screen.press(keypress.Letter('j'))

		typedName(screen, 'r', "s")

		assert.Contains(t, screen.screen(), "3 Root / dockers · by title")
	})

	t.Run("reports a failed create", func(t *testing.T) {
		t.Parallel()

		creator := NewMockFolderCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Folder{}, errDatabaseLocked)
		screen := start(t, folderModel(t, folderActions{
			lister: listerOf(t), trees: treeOf(t, emptyTree()), creator: creator, renamer: NewMockFolderRenamer(t),
		}), wideWidth, wideHeight)

		typedName(screen, 'N', "awk")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("reports a failed rename", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		renamer := NewMockFolderRenamer(t)
		renamer.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Folder{}, errDatabaseLocked)

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: treeOf(t, sample.Tree), creator: NewMockFolderCreator(t), renamer: renamer,
		}), wideWidth, wideHeight)
		screen.press(keypress.Letter('j'))

		typedName(screen, 'r', "s")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func TestModel_folderDelete(t *testing.T) {
	t.Parallel()

	t.Run("confirms with the counts, deletes, and makes the parent the Browse selection", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		previewer := NewMockFolderDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, folder.PreviewDeleteInput{FolderID: sample.Tests.ID()}).
			Return(folder.DeletePreview{Folder: sample.Tests, SubfolderCount: 0, SnippetCount: 2}, nil)

		deleter := NewMockFolderDeleter(t)
		deleter.EXPECT().Run(mock.Anything, folder.DeleteInput{FolderID: sample.Tests.ID()}).Return(nil)

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		listingIn(lister, sample.Go.ID())
		listingIn(lister, sample.Tests.ID())
		listingIn(lister, sample.Go.ID(), filedIn(t, sample.Go.ID()))
		pruned := sample.Tree
		pruned.Folders = []browse.FolderNode{sample.Tree.Folders[0], sample.Tree.Folders[1]}
		pruned.Folders[1].Children = nil
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: treesOf(t, sample.Tree, pruned), previewer: previewer, deleter: deleter,
		}), wideWidth, wideHeight)
		screen.press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('d'))

		assert.Contains(t, screen.screen(),
			`Permanently delete Folder "testing", 0 subfolders, and 2 Snippets? This cannot be undone. [y/N]`)

		screen.press(keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "3 Root / go · by title")
		assert.Contains(t, screen.screen(), filedTitle)
		assert.NotContains(t, screen.screen(), "  testing")
	})

	t.Run("recounts the Tags once the Folder is gone", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		previewer := NewMockFolderDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).
			Return(folder.DeletePreview{Folder: sample.Docker, SubfolderCount: 0, SnippetCount: 3}, nil)

		deleter := NewMockFolderDeleter(t)
		deleter.EXPECT().Run(mock.Anything, mock.Anything).Return(nil)

		counted := sampleTagCounts(t)[0]
		emptied := counted
		emptied.SnippetCount = 0
		tags := NewMockTagLister(t)
		tags.EXPECT().Run(mock.Anything, browse.TagListInput{}).Return([]browse.TagCount{counted}, nil).Once()
		tags.EXPECT().Run(mock.Anything, browse.TagListInput{}).Return([]browse.TagCount{emptied}, nil).Once()

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: treesOf(t, sample.Tree, sample.Tree), previewer: previewer, deleter: deleter,
			tags: tags,
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Regexp(t, `# docker +0`, screen.screen())
	})

	t.Run("reports a failed preview without asking", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		previewer := NewMockFolderDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).Return(folder.DeletePreview{}, errDatabaseLocked)

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: treeOf(t, sample.Tree), previewer: previewer,
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'), keypress.Letter('d'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.NotContains(t, screen.screen(), "Delete Folder")
	})

	t.Run("reports a failed delete", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		previewer := NewMockFolderDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).
			Return(folder.DeletePreview{Folder: sample.Docker, SubfolderCount: 0, SnippetCount: 0}, nil)

		deleter := NewMockFolderDeleter(t)
		deleter.EXPECT().Run(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		lister := listerOf(t)
		listingIn(lister, sample.Docker.ID())
		screen := start(t, folderModel(t, folderActions{
			lister: lister, trees: treeOf(t, sample.Tree), previewer: previewer, deleter: deleter,
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

type folderActions struct {
	lister    tui.FolderSnippetsLister
	trees     tui.FolderTreeLister
	creator   tui.FolderCreator
	renamer   tui.FolderRenamer
	previewer tui.FolderDeletePreviewer
	deleter   tui.FolderDeleter
	tags      tui.TagLister
}

func folderModel(t *testing.T, with folderActions) tui.Model {
	t.Helper()

	return modelWith(t, actions{
		lister:                with.lister,
		treeLister:            with.trees,
		tagLister:             with.tags,
		copier:                NewMockSnippetCopier(t),
		creator:               NewMockSnippetCreator(t),
		searcher:              NewMockSnippetSearcher(t),
		folderCreator:         with.creator,
		folderRenamer:         with.renamer,
		folderDeletePreviewer: with.previewer,
		folderDeleter:         with.deleter,
	})
}

func typedName(screen *driver, opening rune, name string) {
	screen.t.Helper()

	screen.press(keypress.Letter(opening))
	screen.press(keypress.Typed(name)...)
	screen.press(keypress.Special(tea.KeyEnter))
}

func treesOf(t *testing.T, first, then browse.Tree) *MockFolderTreeLister {
	t.Helper()

	lister := NewMockFolderTreeLister(t)
	lister.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).Return(first, nil).Once()
	lister.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).Return(then, nil).Once()

	return lister
}

func withFolderAtRoot(tree browse.Tree, added domain.Folder) browse.Tree {
	tree.Folders = append([]browse.FolderNode{{Folder: added, SnippetCount: 1, Children: nil}}, tree.Folders...)

	return tree
}
