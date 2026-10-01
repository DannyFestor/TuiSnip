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

func TestSnippetRepository_Find(t *testing.T) {
	t.Parallel()

	t.Run("returns an inserted Snippet at the Root", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		snippet := testkit.Snippet(t, testkit.SnippetSpec{
			Title:       "curl json",
			Description: "POST with a JSON body",
			Fragment:    testkit.FragmentSpec{Language: "Bash", Content: "curl -d @body.json\n"},
		})
		require.NoError(t, repository.Insert(t.Context(), snippet))

		got, err := repository.Find(t.Context(), snippet.ID())

		require.NoError(t, err)
		assert.Equal(t, snippet, got)
	})

	t.Run("returns an inserted Snippet inside a Folder", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		repository := newSnippetRepository(t, openDatabase(t, path))
		folderID := testkit.NewSequentialIDs().NewFolderID()
		execRaw(
			t,
			path,
			"INSERT INTO folders (id, name, default_language, created_at, updated_at) VALUES (?, 'scripts', 'Bash', 1, 1)",
			folderID.String(),
		)
		snippet := testkit.Snippet(t, testkit.SnippetSpec{FolderID: folderID})
		require.NoError(t, repository.Insert(t.Context(), snippet))

		got, err := repository.Find(t.Context(), snippet.ID())

		require.NoError(t, err)
		assert.Equal(t, folderID, got.FolderID())
	})

	t.Run("reports an unknown ID as not found", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))

		_, err := repository.Find(t.Context(), testkit.Snippet(t, testkit.SnippetSpec{}).ID())

		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestSnippetRepository_Insert(t *testing.T) {
	t.Parallel()

	t.Run("writes nothing when its Fragment cannot be written", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		sharedFragmentID := ids.NewFragmentID()
		first := testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Fragment: testkit.FragmentSpec{ID: sharedFragmentID},
		})
		second := testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Fragment: testkit.FragmentSpec{ID: sharedFragmentID},
		})
		require.NoError(t, repository.Insert(t.Context(), first))

		require.Error(t, repository.Insert(t.Context(), second))

		_, err := repository.Find(t.Context(), second.ID())
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestSnippetRepository_corruptRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		read func(t *testing.T, repository *sqlite.SnippetRepository, id domain.SnippetID) error
	}{
		{
			name: "Find rejects a Fragment with an unknown Language",
			read: func(t *testing.T, repository *sqlite.SnippetRepository, id domain.SnippetID) error {
				t.Helper()

				_, err := repository.Find(t.Context(), id)

				return err
			},
		},
		{
			name: "List rejects a Fragment with an unknown Language",
			read: func(t *testing.T, repository *sqlite.SnippetRepository, _ domain.SnippetID) error {
				t.Helper()

				_, err := repository.List(t.Context())

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged bytes.Buffer

			path := newDatabasePath(t)
			repository := sqlite.NewSnippetRepository(
				openDatabase(t, path),
				slog.New(slog.NewJSONHandler(&logged, nil)),
			)
			snippet := testkit.Snippet(t, testkit.SnippetSpec{})
			require.NoError(t, repository.Insert(t.Context(), snippet))
			execRaw(t, path, "UPDATE fragments SET language = 'Klingon'")

			err := tt.read(t, repository, snippet.ID())

			require.ErrorIs(t, err, domain.ErrCorruptRecord)
			assert.Contains(t, logged.String(), `"level":"ERROR"`)
			assert.Contains(t, logged.String(), `"snippet_id":"`+snippet.ID().String()+`"`)
		})
	}
}

func TestSnippetRepository_List(t *testing.T) {
	t.Parallel()

	t.Run("returns every inserted Snippet with its Fragment", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		first := testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Title:    "curl json",
			Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID(), Content: "curl -d @body.json"},
		})
		second := testkit.Snippet(t, testkit.SnippetSpec{
			ID:    ids.NewSnippetID(),
			Title: "git undo",
			Fragment: testkit.FragmentSpec{
				ID:       ids.NewFragmentID(),
				Language: "Bash",
				Content:  "git reset HEAD~",
			},
		})
		require.NoError(t, repository.Insert(t.Context(), first))
		require.NoError(t, repository.Insert(t.Context(), second))

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Snippet{first, second}, got)
	})

	t.Run("returns no Snippets from an empty database", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
