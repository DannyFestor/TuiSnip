package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_tags(t *testing.T) {
	t.Parallel()

	t.Run("shows the Tags with their Snippet counts", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		with := taggedActions(t, listerOf(t), tagsOf(t, tags...), NewMockTagSnippetsLister(t))

		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		assert.Regexp(t, `# docker +3`, screen.screen())
		assert.Regexp(t, `# go +4`, screen.screen())
	})

	t.Run("moving the cursor onto a Tag lists the Snippets carrying it", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		carrying := filedIn(t, domain.FolderID{})
		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[1].Tag.ID(), domain.SortOrderTitle, carrying)
		screen := start(
			t,
			modelWith(t, taggedActions(t, listerOf(t, sampleSnippets(t)...), tagsOf(t, tags...), tagged)),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('2'), keypress.Letter('j'))

		assert.Contains(t, screen.screen(), "3 # go · by title")
		assert.Contains(t, screen.screen(), filedTitle)
		assert.NotContains(t, screen.screen(), "Prune everything")
	})

	t.Run("enter on a Tag lists the Snippets carrying it and focuses the Snippet list", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, filedIn(t, domain.FolderID{}))
		screen := start(t, modelWith(t, taggedActions(t, listerOf(t), tagsOf(t, tags...), tagged)),
			wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.screen(), "3 # docker · by title")
		assert.Contains(t, screen.screen(), filedTitle)
	})

	t.Run("cycling the sort order relists the Tag's Snippets in the next order", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[1].Tag.ID(), domain.SortOrderTitle, filedIn(t, domain.FolderID{}))
		listingWithTag(tagged, tags[1].Tag.ID(), domain.SortOrderUpdated, filedIn(t, domain.FolderID{}))
		with := taggedActions(t, listerOf(t), tagsOf(t, tags...), tagged)
		saver := NewMockSortOrderSaver(t)
		saver.EXPECT().SaveSortOrder(mock.Anything, domain.SortOrderUpdated).Return(nil).Once()
		with.sortOrderSaver = saver
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Typed("2j3s")...)

		assert.Contains(t, screen.screen(), "3 # go · by last updated")
	})

	t.Run("reports a Tag list failure", func(t *testing.T) {
		t.Parallel()

		tags := NewMockTagLister(t)
		tags.EXPECT().Run(mock.Anything, browse.TagListInput{}).Return(nil, errDatabaseLocked)

		screen := start(t, modelWith(t, taggedActions(t, listerOf(t), tags, NewMockTagSnippetsLister(t))),
			wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("reports a failure to list a Tag's Snippets", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		tagged := NewMockTagSnippetsLister(t)
		tagged.EXPECT().
			Run(mock.Anything, browse.SnippetsWithTagInput{TagID: tags[1].Tag.ID(), Order: domain.SortOrderTitle}).
			Return(nil, errDatabaseLocked)
		screen := start(t, modelWith(t, taggedActions(t, listerOf(t), tagsOf(t, tags...), tagged)),
			wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Letter('j'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func taggedActions(
	t *testing.T, lister *MockFolderSnippetsLister, tags *MockTagLister, tagged *MockTagSnippetsLister,
) actions {
	t.Helper()

	return actions{
		lister:            lister,
		treeLister:        treeOf(t, emptyTree()),
		tagLister:         tags,
		tagSnippetsLister: tagged,
		copier:            NewMockSnippetCopier(t),
		creator:           NewMockSnippetCreator(t),
		searcher:          NewMockSnippetSearcher(t),
	}
}

func sampleTagCounts(t *testing.T) []browse.TagCount {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return []browse.TagCount{
		{Tag: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"}), SnippetCount: 3},
		{Tag: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}), SnippetCount: 4},
	}
}

func listingWithTag(
	lister *MockTagSnippetsLister, tagID domain.TagID, order domain.SortOrder, snippets ...domain.Snippet,
) {
	lister.EXPECT().
		Run(mock.Anything, browse.SnippetsWithTagInput{TagID: tagID, Order: order}).
		Return(snippets, nil).
		Once()
}
