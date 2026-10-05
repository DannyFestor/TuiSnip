package sqlite_test

import (
	"bytes"
	"testing"
	"time"

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
		insertRawFolder(t, path, folderID)
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

	t.Run("Find rejects a Fragment with an unknown Language", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newLoggingSnippetRepository(t, path, &logged)
		snippet := insertCorruptSnippet(t, repository, path, testkit.NewSequentialIDs())

		_, err := repository.Find(t.Context(), snippet.ID())

		require.ErrorIs(t, err, domain.ErrCorruptRecord)
		assertCorruptRowLogged(t, &logged, "ERROR", snippet.ID())
	})

	tests := []struct {
		name string
		list func(t *testing.T, repository *sqlite.SnippetRepository) ([]domain.Snippet, error)
	}{
		{
			name: "List skips a Snippet whose Fragment has an unknown Language",
			list: func(t *testing.T, repository *sqlite.SnippetRepository) ([]domain.Snippet, error) {
				t.Helper()

				return repository.List(t.Context())
			},
		},
		{
			name: "ListInFolder skips a Snippet whose Fragment has an unknown Language",
			list: func(t *testing.T, repository *sqlite.SnippetRepository) ([]domain.Snippet, error) {
				t.Helper()

				return repository.ListInFolder(t.Context(), domain.FolderID{}, domain.SortOrderTitle)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged bytes.Buffer

			path := newDatabasePath(t)
			repository := newLoggingSnippetRepository(t, path, &logged)
			ids := testkit.NewSequentialIDs()
			healthy := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "healthy"})
			corrupt := insertCorruptSnippet(t, repository, path, ids)

			got, err := tt.list(t, repository)

			require.NoError(t, err)
			assert.Equal(t, []domain.Snippet{healthy}, got)
			assertCorruptRowLogged(t, &logged, "WARN", corrupt.ID())
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

func TestSnippetRepository_ListInFolder(t *testing.T) {
	t.Parallel()

	t.Run("returns the Snippets at the Root by title, ignoring case", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		repository := newSnippetRepository(t, openDatabase(t, path))
		ids := testkit.NewSequentialIDs()
		folderID := ids.NewFolderID()
		insertRawFolder(t, path, folderID)
		zebra := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "zebra"})
		apple := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "Apple"})
		insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "filed", FolderID: folderID})

		got, err := repository.ListInFolder(t.Context(), domain.FolderID{}, domain.SortOrderTitle)

		require.NoError(t, err)
		assert.Equal(t, []domain.Snippet{apple, zebra}, got)
	})

	for _, order := range []domain.SortOrder{domain.SortOrderTitle, domain.SortOrderUpdated, domain.SortOrderCreated} {
		t.Run("returns only the Snippets inside the Folder by "+order.String(), func(t *testing.T) {
			t.Parallel()

			path := newDatabasePath(t)
			repository := newSnippetRepository(t, openDatabase(t, path))
			ids := testkit.NewSequentialIDs()
			folderID := ids.NewFolderID()
			insertRawFolder(t, path, folderID)
			insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "at the Root"})
			filed := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "filed", FolderID: folderID})

			got, err := repository.ListInFolder(t.Context(), folderID, order)

			require.NoError(t, err)
			assert.Equal(t, []domain.Snippet{filed}, got)
		})
	}

	t.Run("returns no Snippets for an empty Folder", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))

		got, err := repository.ListInFolder(t.Context(), domain.FolderID{}, domain.SortOrderTitle)

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestSnippetRepository_ListInFolder_SortOrder(t *testing.T) {
	t.Parallel()

	earlier := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
	later := earlier.Add(time.Hour)
	latest := later.Add(time.Hour)

	tests := []struct {
		name  string
		order domain.SortOrder
		specs []testkit.SnippetSpec
		want  []string
	}{
		{
			name:  "by last updated lists the most recently updated first",
			order: domain.SortOrderUpdated,
			specs: []testkit.SnippetSpec{
				{Title: "a", CreatedAt: earlier, UpdatedAt: later.Add(time.Minute)},
				{Title: "b", CreatedAt: earlier, UpdatedAt: latest},
				{Title: "c", CreatedAt: later, UpdatedAt: later},
			},
			want: []string{"b", "a", "c"},
		},
		{
			name:  "by creation date lists the newest first",
			order: domain.SortOrderCreated,
			specs: []testkit.SnippetSpec{
				{Title: "a", CreatedAt: earlier, UpdatedAt: latest},
				{Title: "b", CreatedAt: later, UpdatedAt: later},
				{Title: "c", CreatedAt: latest, UpdatedAt: latest},
			},
			want: []string{"c", "b", "a"},
		},
		{
			name:  "breaks a tie on last updated by title, ignoring case",
			order: domain.SortOrderUpdated,
			specs: []testkit.SnippetSpec{
				{Title: "zebra", CreatedAt: earlier},
				{Title: "Apple", CreatedAt: earlier},
			},
			want: []string{"Apple", "zebra"},
		},
		{
			name:  "breaks a tie on creation date by title, ignoring case",
			order: domain.SortOrderCreated,
			specs: []testkit.SnippetSpec{
				{Title: "zebra", CreatedAt: earlier},
				{Title: "Apple", CreatedAt: earlier},
			},
			want: []string{"Apple", "zebra"},
		},
		{
			name:  "breaks a tie on title by ID",
			order: domain.SortOrderTitle,
			specs: []testkit.SnippetSpec{
				{Title: "same"},
				{Title: "SAME"},
			},
			want: []string{"same", "SAME"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
			ids := testkit.NewSequentialIDs()

			for _, spec := range tt.specs {
				insertSnippet(t, repository, ids, spec)
			}

			got, err := repository.ListInFolder(t.Context(), domain.FolderID{}, tt.order)

			require.NoError(t, err)
			assert.Equal(t, tt.want, titlesOf(got))
		})
	}
}

func titlesOf(snippets []domain.Snippet) []string {
	titles := make([]string, 0, len(snippets))
	for index := range snippets {
		titles = append(titles, snippets[index].Title().String())
	}

	return titles
}

func TestSnippetRepository_CountByFolder(t *testing.T) {
	t.Parallel()

	t.Run("counts the Snippets in each Folder and at the Root", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		repository := newSnippetRepository(t, openDatabase(t, path))
		ids := testkit.NewSequentialIDs()
		folderID := ids.NewFolderID()
		insertRawFolder(t, path, folderID)
		insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "at the Root"})
		insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "first filed", FolderID: folderID})
		insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "second filed", FolderID: folderID})

		got, err := repository.CountByFolder(t.Context())

		require.NoError(t, err)
		assert.Equal(t, map[domain.FolderID]int{{}: 1, folderID: 2}, got)
	})

	t.Run("counts nothing in an empty database", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))

		got, err := repository.CountByFolder(t.Context())

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func assertCorruptRowLogged(t *testing.T, logged *bytes.Buffer, level string, id domain.SnippetID) {
	t.Helper()

	assert.Contains(t, logged.String(), `"level":"`+level+`"`)
	assert.Contains(t, logged.String(), `"snippet_id":"`+id.String()+`"`)
	assert.Contains(t, logged.String(), `"error":"`+domain.ErrCorruptRecord.Error())
}
