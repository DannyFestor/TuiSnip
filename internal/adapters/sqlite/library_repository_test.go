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

func TestLibraryRepository_InsertIntoEmpty(t *testing.T) {
	t.Parallel()

	t.Run("inserts every Folder, Tag, and Snippet, a subfolder listed before its parent included", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		ids := testkit.NewSequentialIDs()
		parentID := ids.NewFolderID()
		child := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "go", ParentID: parentID})
		parent := testkit.Folder(t, testkit.FolderSpec{ID: parentID, Name: "scripts"})
		carried := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
		unused := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "unused"})
		filed := testkit.Snippet(t, testkit.SnippetSpec{
			ID: ids.NewSnippetID(), FolderID: child.ID(), Fragments: severalFragments(ids), Tags: []domain.Tag{carried},
		})
		atRoot := testkit.Snippet(t, testkit.SnippetSpec{
			ID: ids.NewSnippetID(), Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID()}},
		})
		repository := newLibraryRepository(t, database, slog.New(slog.DiscardHandler))

		err := repository.InsertIntoEmpty(t.Context(), domain.Library{
			Folders:  []domain.Folder{child, parent},
			Tags:     []domain.Tag{carried, unused},
			Snippets: []domain.Snippet{filed, atRoot},
		})

		require.NoError(t, err)
		got, err := repository.Load(t.Context())
		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Folder{parent, child}, got.Folders)
		assert.ElementsMatch(t, []domain.Tag{carried, unused}, got.Tags)
		assert.ElementsMatch(t, []domain.Snippet{filed, atRoot}, got.Snippets)
	})

	refusals := []struct {
		name string
		seed func(t *testing.T, database *sqlite.Database)
	}{
		{name: "refuses a Library holding a Folder and writes nothing", seed: seedFolder},
		{name: "refuses a Library holding a Tag and writes nothing", seed: seedTag},
		{name: "refuses a Library holding a Snippet and writes nothing", seed: seedSnippet},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			database := openDatabase(t, newDatabasePath(t))
			tt.seed(t, database)
			repository := newLibraryRepository(t, database, slog.New(slog.DiscardHandler))
			before, err := repository.Load(t.Context())
			require.NoError(t, err)

			err = repository.InsertIntoEmpty(t.Context(), importedLibrary(t))

			require.ErrorIs(t, err, domain.ErrLibraryNotEmpty)
			require.ErrorContains(t, err, "sqlite.LibraryRepository.InsertIntoEmpty: ")
			after, err := repository.Load(t.Context())
			require.NoError(t, err)
			assert.Equal(t, before, after)
		})
	}

	t.Run("writes nothing when a record fails to insert", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		ids := testkit.NewSequentialIDs()
		repository := newLibraryRepository(t, database, slog.New(slog.DiscardHandler))

		err := repository.InsertIntoEmpty(t.Context(), domain.Library{
			Folders: []domain.Folder{testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})},
			Tags: []domain.Tag{
				testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Shell"}),
				testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"}),
			},
			Snippets: nil,
		})

		require.ErrorIs(t, err, domain.ErrTagNameTaken)
		got, err := repository.Load(t.Context())
		require.NoError(t, err)
		assert.Empty(t, got.Folders)
		assert.Empty(t, got.Tags)
	})
}

func TestLibraryRepository_InsertAlongside(t *testing.T) {
	t.Parallel()

	t.Run("adds beside the stored records, reusing a stored Tag of the same name", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		ids := testkit.NewSequentialIDs()
		storedFolder := insertFolder(
			t,
			newFolderRepository(t, database, slog.New(slog.DiscardHandler)),
			testkit.FolderSpec{
				ID: ids.NewFolderID(), Name: "scripts",
			},
		)
		storedTag := insertTag(t, newTagRepository(t, database), ids, "go")
		folder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "scripts"})
		sameName := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Go"})
		fresh := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
		snippet := testkit.Snippet(t, testkit.SnippetSpec{
			ID:        ids.NewSnippetID(),
			FolderID:  folder.ID(),
			Fragments: severalFragments(ids),
			Tags:      []domain.Tag{sameName, fresh},
		})
		repository := newLibraryRepository(t, database, slog.New(slog.DiscardHandler))

		written, err := repository.InsertAlongside(t.Context(), domain.Library{
			Folders:  []domain.Folder{folder},
			Tags:     []domain.Tag{sameName, fresh},
			Snippets: []domain.Snippet{snippet},
		})

		require.NoError(t, err)
		assert.Equal(t, []domain.Tag{fresh}, written.Tags)
		got, err := repository.Load(t.Context())
		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Folder{storedFolder, folder}, got.Folders)
		assert.ElementsMatch(t, []domain.Tag{storedTag, fresh}, got.Tags)
		require.Len(t, got.Snippets, 1)
		assert.Equal(t, []domain.Tag{storedTag, fresh}, got.Snippets[0].Tags())
	})

	t.Run("writes nothing when a record fails to insert", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		ids := testkit.NewSequentialIDs()
		insertTag(t, newTagRepository(t, database), ids, "go")
		repository := newLibraryRepository(t, database, slog.New(slog.DiscardHandler))
		before, err := repository.Load(t.Context())
		require.NoError(t, err)
		twin := testkit.Snippet(t, testkit.SnippetSpec{
			ID: ids.NewSnippetID(), Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID()}},
		})

		_, err = repository.InsertAlongside(t.Context(), domain.Library{
			Folders:  []domain.Folder{testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID()})},
			Tags:     []domain.Tag{testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})},
			Snippets: []domain.Snippet{twin, twin},
		})

		require.ErrorContains(t, err, "sqlite.LibraryRepository.InsertAlongside: ")
		after, err := repository.Load(t.Context())
		require.NoError(t, err)
		assert.Equal(t, before, after)
	})
}

func importedLibrary(t *testing.T) domain.Library {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return domain.Library{
		Folders:  []domain.Folder{testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "imported"})},
		Tags:     []domain.Tag{testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "imported"})},
		Snippets: nil,
	}
}

func seedFolder(t *testing.T, database *sqlite.Database) {
	t.Helper()

	insertFolder(t, newFolderRepository(t, database, slog.New(slog.DiscardHandler)), testkit.FolderSpec{
		ID: testkit.NewSequentialIDs().NewFolderID(), Name: "seeded",
	})
}

func seedTag(t *testing.T, database *sqlite.Database) {
	t.Helper()

	insertTag(t, newTagRepository(t, database), testkit.NewSequentialIDs(), "seeded")
}

func seedSnippet(t *testing.T, database *sqlite.Database) {
	t.Helper()

	insertSnippet(
		t,
		newSnippetRepository(t, database),
		testkit.NewSequentialIDs(),
		testkit.SnippetSpec{Title: "seeded"},
	)
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
