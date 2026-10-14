package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestLibrary_InCreatedOrder(t *testing.T) {
	t.Parallel()

	t.Run("orders Folders by created date, then ID", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		lowID := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), CreatedAt: later()})
		oldest := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), CreatedAt: earlier()})
		highID := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), CreatedAt: later()})

		got := domain.Library{Folders: []domain.Folder{highID, lowID, oldest}}.InCreatedOrder()

		assert.Equal(t, []domain.Folder{oldest, lowID, highID}, got.Folders)
	})

	t.Run("orders Tags by created date, then ID", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		lowID := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), CreatedAt: later()})
		oldest := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), CreatedAt: earlier()})
		highID := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), CreatedAt: later()})

		got := domain.Library{Tags: []domain.Tag{highID, lowID, oldest}}.InCreatedOrder()

		assert.Equal(t, []domain.Tag{oldest, lowID, highID}, got.Tags)
	})

	t.Run("orders Snippets by created date, then ID", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		lowID := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: later()})
		oldest := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: earlier()})
		highID := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: later()})

		got := domain.Library{Snippets: []domain.Snippet{highID, lowID, oldest}}.InCreatedOrder()

		assert.Equal(t, []domain.Snippet{oldest, lowID, highID}, got.Snippets)
	})

	t.Run("leaves the original order alone", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		newer := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), CreatedAt: later()})
		older := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), CreatedAt: earlier()})
		library := domain.Library{Tags: []domain.Tag{newer, older}}

		library.InCreatedOrder()

		assert.Equal(t, []domain.Tag{newer, older}, library.Tags)
	})
}

func TestLibrary_FoldersParentsFirst(t *testing.T) {
	t.Parallel()

	t.Run("puts every parent before its subfolders and keeps the order otherwise", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		grandparentID, parentID := ids.NewFolderID(), ids.NewFolderID()
		child := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), ParentID: parentID})
		parent := testkit.Folder(t, testkit.FolderSpec{ID: parentID, ParentID: grandparentID})
		atRoot := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
		grandparent := testkit.Folder(t, testkit.FolderSpec{ID: grandparentID})

		got, err := domain.Library{Folders: []domain.Folder{child, parent, atRoot, grandparent}}.FoldersParentsFirst()

		require.NoError(t, err)
		assert.Equal(t, []domain.Folder{atRoot, grandparent, parent, child}, got)
	})

	t.Run("refuses Folders whose parents loop and names each one", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		firstID, secondID := ids.NewFolderID(), ids.NewFolderID()
		first := testkit.Folder(t, testkit.FolderSpec{ID: firstID, ParentID: secondID})
		second := testkit.Folder(t, testkit.FolderSpec{ID: secondID, ParentID: firstID})

		_, err := domain.Library{Folders: []domain.Folder{first, second}}.FoldersParentsFirst()

		require.ErrorIs(t, err, domain.ErrFolderCycle)
		require.ErrorContains(t, err, "folder "+firstID.String()+": ")
		assert.ErrorContains(t, err, "folder "+secondID.String()+": ")
	})

	t.Run("refuses a Folder whose parent the Library lacks and names both", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		missingID := ids.NewFolderID()
		orphan := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), ParentID: missingID})

		_, err := domain.Library{Folders: []domain.Folder{orphan}}.FoldersParentsFirst()

		require.ErrorIs(t, err, domain.ErrMissingFolder)
		assert.ErrorContains(t, err, "folder "+orphan.ID().String()+": parent "+missingID.String()+": ")
	})
}

func TestLibrary_RequireIntact(t *testing.T) {
	t.Parallel()

	t.Run("accepts Snippets filed in the Library's Folders and carrying its Tags", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		folder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID()})
		filed := testkit.Snippet(t, testkit.SnippetSpec{
			ID: ids.NewSnippetID(), FolderID: folder.ID(), Tags: []domain.Tag{tag},
		})
		atRoot := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID()})

		err := domain.Library{
			Folders:  []domain.Folder{folder},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{filed, atRoot},
		}.RequireIntact()

		assert.NoError(t, err)
	})

	t.Run("refuses a Snippet in a Folder the Library lacks and names both", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		missingID := ids.NewFolderID()
		snippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), FolderID: missingID})

		err := domain.Library{Snippets: []domain.Snippet{snippet}}.RequireIntact()

		require.ErrorIs(t, err, domain.ErrMissingFolder)
		assert.ErrorContains(t, err, "snippet "+snippet.ID().String()+": folder "+missingID.String()+": ")
	})

	t.Run("refuses a Snippet carrying a Tag the Library lacks and names both", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		missing := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID()})
		snippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), Tags: []domain.Tag{missing}})

		err := domain.Library{Snippets: []domain.Snippet{snippet}}.RequireIntact()

		require.ErrorIs(t, err, domain.ErrMissingTag)
		assert.ErrorContains(t, err, "snippet "+snippet.ID().String()+": tag "+missing.ID().String()+": ")
	})

	t.Run("refuses Tags whose names clash ignoring case and names the later one", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		first := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Shell"})
		clashing := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})

		err := domain.Library{Tags: []domain.Tag{first, clashing}}.RequireIntact()

		require.ErrorIs(t, err, domain.ErrTagNameTaken)
		require.ErrorContains(t, err, "tag "+clashing.ID().String()+": ")
		assert.NotContains(t, err.Error(), "tag "+first.ID().String()+": ")
	})

	t.Run("refuses Folders whose parents loop", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		firstID, secondID := ids.NewFolderID(), ids.NewFolderID()
		first := testkit.Folder(t, testkit.FolderSpec{ID: firstID, ParentID: secondID})
		second := testkit.Folder(t, testkit.FolderSpec{ID: secondID, ParentID: firstID})

		err := domain.Library{Folders: []domain.Folder{first, second}}.RequireIntact()

		assert.ErrorIs(t, err, domain.ErrFolderCycle)
	})
}

func earlier() time.Time {
	return time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
}

func later() time.Time {
	return time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)
}
