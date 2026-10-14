package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

func earlier() time.Time {
	return time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
}

func later() time.Time {
	return time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)
}
