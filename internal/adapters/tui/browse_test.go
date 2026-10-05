package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_browse(t *testing.T) {
	t.Parallel()

	t.Run("shows the Folder tree with its Snippet counts", func(t *testing.T) {
		t.Parallel()

		screen := start(t, browsingModel(t, foldertree.New(t), listerOf(t)), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "▾ go")
		assert.Regexp(t, `testing +1`, screen.screen())
	})

	t.Run("moving the cursor onto a Folder lists its Snippets", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		lister := listerOf(t, sampleSnippets(t)...)
		listingIn(lister, sample.Docker.ID(), filedIn(t, sample.Docker.ID()))
		screen := start(t, browsingModel(t, sample, lister), wideWidth, wideHeight)

		screen.press(keypress.Letter('j'))

		assert.Contains(t, screen.screen(), "3 Root / docker · by title")
		assert.Contains(t, screen.screen(), filedTitle)
		assert.NotContains(t, screen.screen(), "Prune everything")
	})

	t.Run("reports a Folder tree failure", func(t *testing.T) {
		t.Parallel()

		tree := NewMockFolderTreeLister(t)
		tree.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).Return(browse.Tree{}, errDatabaseLocked)
		screen := start(t, modelWith(t, actions{
			lister:     listerOf(t),
			treeLister: tree,
			copier:     NewMockSnippetCopier(t),
			creator:    NewMockSnippetCreator(t),
			searcher:   NewMockSnippetSearcher(t),
		}), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("revealing a Search hit in a Folder selects that Folder and that Snippet", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		filed := filedIn(t, sample.Tests.ID())
		lister := listerOf(t, sampleSnippets(t)...)
		listingIn(lister, sample.Tests.ID(), sampleSnippets(t)[0], filed)
		searcher := NewMockSnippetSearcher(t)
		searcher.EXPECT().Run(mock.Anything, search.QueryInput{Text: "t"}).Return(hitsOf(filed), nil)
		screen := start(t, modelWith(t, actions{
			lister:     lister,
			treeLister: treeOf(t, sample.Tree),
			copier:     NewMockSnippetCopier(t),
			creator:    NewMockSnippetCreator(t),
			searcher:   searcher,
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('/'), keypress.Letter('t'), keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.screen(), "3 Root / go / testing ")
		assert.Contains(t, screen.screen(), "Root / go / testing · Go")
		assert.Contains(t, screen.screen(), "func TestParse")
	})

	t.Run("reloads the Folder tree after a save", func(t *testing.T) {
		t.Parallel()

		saved := numberedSnippets(t, 1)[0]
		tree := NewMockFolderTreeLister(t)
		tree.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).Return(emptyTree(), nil).Once()
		tree.EXPECT().Run(mock.Anything, browse.FolderTreeInput{}).
			Return(browse.Tree{RootSnippetCount: 1, Folders: nil}, nil).Once()

		creator := NewMockSnippetCreator(t)
		creator.EXPECT().Run(mock.Anything, snippet.CreateInput{
			Title: "Prune", Description: "", Language: "plaintext", Content: "",
		}).Return(saved, nil)
		screen := start(t, modelWith(t, actions{
			lister:     listerReturning(t, nil, []domain.Snippet{saved}),
			treeLister: tree,
			copier:     NewMockSnippetCopier(t),
			creator:    creator,
			searcher:   NewMockSnippetSearcher(t),
		}), wideWidth, wideHeight)

		screen.press(keypress.Letter('n'))
		screen.press(keypress.Typed("Prune")...)
		screen.press(keypress.Ctrl('s'))

		assert.Regexp(t, `◆ Root +1`, screen.screen())
	})
}
