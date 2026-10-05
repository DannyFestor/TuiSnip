package tui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
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

func TestModel_tagNaming(t *testing.T) {
	t.Parallel()

	t.Run("creates a typed Tag and makes it the Browse selection", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		created := testkit.Tag(t, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "awk"})
		creator := NewMockTagCreator(t)
		creator.EXPECT().Run(mock.Anything, tag.CreateInput{Name: "awk"}).Return(created, nil)

		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, created.ID(), domain.SortOrderTitle)
		with := taggedActions(t, listerOf(t), tagListsOf(t, tags, withTag(tags, created, 0)), tagged)
		with.tagCreator = creator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'))
		typedName(screen, 'N', "awk")

		assert.Contains(t, screen.screen(), "3 # awk · by title")
		assert.Regexp(t, `# awk +0`, screen.screen())
	})

	t.Run("refuses a name a Tag has without creating it", func(t *testing.T) {
		t.Parallel()

		with := taggedActions(t, listerOf(t), tagsOf(t, sampleTagCounts(t)...), NewMockTagSnippetsLister(t))
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'))
		typedName(screen, 'N', "Go")

		assert.Contains(t, screen.screen(), "Tag go already exists")
	})

	t.Run("renames the Tag under the cursor onto another's name and lists the merged Tag", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		survivor := testkit.Tag(t, testkit.TagSpec{ID: tags[0].Tag.ID(), Name: "Docker"})
		renamer := NewMockTagRenamer(t)
		renamer.EXPECT().Run(mock.Anything, tag.RenameInput{TagID: tags[1].Tag.ID(), Name: "Docker"}).
			Return(survivor, nil)

		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[1].Tag.ID(), domain.SortOrderTitle, filedIn(t, domain.FolderID{}))
		listingWithTag(tagged, survivor.ID(), domain.SortOrderTitle, filedIn(t, domain.FolderID{}))
		merged := []browse.TagCount{{Tag: survivor, SnippetCount: 6}}
		with := taggedActions(t, listerOf(t), tagListsOf(t, tags, merged), tagged)
		with.tagRenamer = renamer
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)
		screen.press(keypress.Letter('2'), keypress.Letter('j'), keypress.Letter('r'))
		screen.press(keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyBackspace))

		screen.press(keypress.Typed("Docker")...)
		screen.press(keypress.Special(tea.KeyEnter))

		assert.Regexp(t, `# Docker +6`, screen.screen())
		assert.Contains(t, screen.screen(), "3 # Docker · by title")
		assert.Contains(t, screen.screen(), filedTitle)
	})

	t.Run("reports a failed create", func(t *testing.T) {
		t.Parallel()

		creator := NewMockTagCreator(t)
		creator.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Tag{}, errDatabaseLocked)

		with := taggedActions(t, listerOf(t), noTags(t), NewMockTagSnippetsLister(t))
		with.tagCreator = creator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'))
		typedName(screen, 'N', "awk")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})

	t.Run("reports a failed rename", func(t *testing.T) {
		t.Parallel()

		renamer := NewMockTagRenamer(t)
		renamer.EXPECT().Run(mock.Anything, mock.Anything).Return(domain.Tag{}, errDatabaseLocked)

		with := taggedActions(t, listerOf(t), tagsOf(t, sampleTagCounts(t)...), NewMockTagSnippetsLister(t))
		with.tagRenamer = renamer
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'))
		typedName(screen, 'r', "s")

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func TestModel_tagDelete(t *testing.T) {
	t.Parallel()

	t.Run("confirms with the count, deletes, and lists the Tag that takes its place", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		previewer := NewMockTagDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, tag.PreviewDeleteInput{TagID: tags[1].Tag.ID()}).
			Return(tag.DeletePreview{Tag: tags[1].Tag, SnippetCount: 4}, nil)

		deleter := NewMockTagDeleter(t)
		deleter.EXPECT().Run(mock.Anything, tag.DeleteInput{TagID: tags[1].Tag.ID()}).Return(nil)

		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[1].Tag.ID(), domain.SortOrderTitle)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, filedIn(t, domain.FolderID{}))
		with := taggedActions(t, listerOf(t), tagListsOf(t, tags, tags[:1]), tagged)
		with.tagDeletePreviewer = previewer
		with.tagDeleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)
		screen.press(keypress.Letter('2'), keypress.Letter('j'), keypress.Letter('d'))

		assert.Contains(t, screen.screen(), `"go"`)
		assert.Contains(t, screen.screen(), "4 Snippets")

		screen.press(keypress.Letter('y'))

		assert.NotRegexp(t, `# go +4`, screen.screen())
		assert.Contains(t, screen.screen(), "3 # docker · by title")
		assert.Contains(t, screen.screen(), filedTitle)
	})

	t.Run("deletes a Tag no Snippet carries without asking", func(t *testing.T) {
		t.Parallel()

		unused := testkit.Tag(t, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "unused"})
		previewer := NewMockTagDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).
			Return(tag.DeletePreview{Tag: unused, SnippetCount: 0}, nil)

		deleter := NewMockTagDeleter(t)
		deleter.EXPECT().Run(mock.Anything, tag.DeleteInput{TagID: unused.ID()}).Return(nil)

		with := taggedActions(
			t, listerOf(t), tagListsOf(t, []browse.TagCount{{Tag: unused, SnippetCount: 0}}, nil),
			NewMockTagSnippetsLister(t),
		)
		with.tagDeletePreviewer = previewer
		with.tagDeleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Letter('d'))

		assert.NotContains(t, screen.screen(), "[y/N]")
		assert.Contains(t, screen.screen(), "No Tags yet.")
	})

	t.Run("reports a failed preview without asking", func(t *testing.T) {
		t.Parallel()

		previewer := NewMockTagDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).Return(tag.DeletePreview{}, errDatabaseLocked)

		with := taggedActions(t, listerOf(t), tagsOf(t, sampleTagCounts(t)...), NewMockTagSnippetsLister(t))
		with.tagDeletePreviewer = previewer
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Letter('d'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
		assert.NotContains(t, screen.screen(), "[y/N]")
	})

	t.Run("reports a failed delete", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		previewer := NewMockTagDeletePreviewer(t)
		previewer.EXPECT().Run(mock.Anything, mock.Anything).
			Return(tag.DeletePreview{Tag: tags[0].Tag, SnippetCount: 3}, nil)

		deleter := NewMockTagDeleter(t)
		deleter.EXPECT().Run(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		with := taggedActions(t, listerOf(t), tagsOf(t, tags...), NewMockTagSnippetsLister(t))
		with.tagDeletePreviewer = previewer
		with.tagDeleter = deleter
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Letter('d'), keypress.Letter('y'))

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func tagListsOf(t *testing.T, first, then []browse.TagCount) *MockTagLister {
	t.Helper()

	lister := NewMockTagLister(t)
	lister.EXPECT().Run(mock.Anything, browse.TagListInput{}).Return(first, nil).Once()
	lister.EXPECT().Run(mock.Anything, browse.TagListInput{}).Return(then, nil).Once()

	return lister
}

func withTag(tags []browse.TagCount, added domain.Tag, count int) []browse.TagCount {
	return append([]browse.TagCount{{Tag: added, SnippetCount: count}}, tags...)
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
