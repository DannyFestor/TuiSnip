package sqlite_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestLibraryRepository_Load(t *testing.T) {
	t.Parallel()

	t.Run("loads every Folder, Tag, and Snippet, Tags no Snippet carries included", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		ids := testkit.NewSequentialIDs()
		folder := insertFolder(t, newFolderRepository(t, database, slog.New(slog.DiscardHandler)), testkit.FolderSpec{
			ID:              ids.NewFolderID(),
			Name:            "scripts",
			DefaultLanguage: "Bash",
		})
		tags := newTagRepository(t, database)
		carried := insertTag(t, tags, ids, "shell")
		unused := insertTag(t, tags, ids, "unused")
		snippets := newSnippetRepository(t, database)
		filed := testkit.Snippet(t, testkit.SnippetSpec{
			ID:        ids.NewSnippetID(),
			FolderID:  folder.ID(),
			Fragments: severalFragments(ids),
			Tags:      []domain.Tag{carried},
		})
		require.NoError(t, snippets.Insert(t.Context(), filed))
		atRoot := insertSnippet(t, snippets, ids, testkit.SnippetSpec{Title: "at the Root"})

		got, err := newLibraryRepository(t, database, slog.New(slog.DiscardHandler)).Load(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []domain.Folder{folder}, got.Folders)
		assert.ElementsMatch(t, []domain.Tag{carried, unused}, got.Tags)
		assert.ElementsMatch(t, []domain.Snippet{filed, atRoot}, got.Snippets)
	})

	t.Run("loads an empty Library", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))

		got, err := newLibraryRepository(t, database, slog.New(slog.DiscardHandler)).Load(t.Context())

		require.NoError(t, err)
		assert.Empty(t, got.Folders)
		assert.Empty(t, got.Tags)
		assert.Empty(t, got.Snippets)
	})

	tests := []struct {
		name     string
		corrupt  func(t *testing.T, path string, database *sqlite.Database) string
		wantKind string
		wantKey  string
	}{
		{
			name:     "fails on a corrupt Folder and names it",
			corrupt:  corruptFolder,
			wantKind: "folder",
			wantKey:  "folder_id",
		},
		{name: "fails on a corrupt Tag and names it", corrupt: corruptTag, wantKind: "tag", wantKey: "tag_id"},
		{
			name:     "fails on a corrupt Snippet and names it",
			corrupt:  corruptSnippet,
			wantKind: "snippet",
			wantKey:  "snippet_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged bytes.Buffer

			path := newDatabasePath(t)
			database := openDatabase(t, path)
			corruptID := tt.corrupt(t, path, database)

			_, err := newLibraryRepository(t, database, slog.New(slog.NewJSONHandler(&logged, nil))).Load(t.Context())

			require.ErrorIs(t, err, domain.ErrCorruptRecord)
			require.ErrorContains(t, err, tt.wantKind+" "+corruptID+": ")
			assert.Contains(t, logged.String(), `"level":"ERROR"`)
			assert.Contains(t, logged.String(), `"`+tt.wantKey+`":"`+corruptID+`"`)
		})
	}
}

func newLibraryRepository(t *testing.T, database *sqlite.Database, logger *slog.Logger) *sqlite.LibraryRepository {
	t.Helper()

	return sqlite.NewLibraryRepository(database, logger)
}

func corruptFolder(t *testing.T, path string, database *sqlite.Database) string {
	t.Helper()

	repository := newFolderRepository(t, database, slog.New(slog.DiscardHandler))
	folder := insertFolder(t, repository, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID()})
	execRaw(t, path, "UPDATE folders SET default_language = 'Klingon' WHERE id = ?", folder.ID().String())

	return folder.ID().String()
}

func corruptTag(t *testing.T, path string, database *sqlite.Database) string {
	t.Helper()

	tag := insertTag(t, newTagRepository(t, database), testkit.NewSequentialIDs(), "shell")
	execRaw(t, path, "UPDATE tags SET name = 'shell,bash' WHERE id = ?", tag.ID().String())

	return tag.ID().String()
}

func corruptSnippet(t *testing.T, path string, database *sqlite.Database) string {
	t.Helper()

	snippet := insertCorruptSnippet(t, newSnippetRepository(t, database), path, testkit.NewSequentialIDs())

	return snippet.ID().String()
}
