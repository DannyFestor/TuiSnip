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

	t.Run("returns an inserted Snippet with the Tags it carries", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		repository := newSnippetRepository(t, database)
		tags := newTagRepository(t, database)
		ids := testkit.NewSequentialIDs()
		golang := insertTag(t, tags, ids, "go")
		docker := insertTag(t, tags, ids, "docker")
		insertTag(t, tags, ids, "unused")
		snippet := insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, docker}})

		got, err := repository.Find(t.Context(), snippet.ID())

		require.NoError(t, err)
		assert.Equal(t, []domain.Tag{docker, golang}, got.Tags())
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

	t.Run("writes nothing when it carries a Tag that was never stored", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		unstored := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID()})
		snippet := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), Tags: []domain.Tag{unstored}})

		require.Error(t, repository.Insert(t.Context(), snippet))

		_, err := repository.Find(t.Context(), snippet.ID())
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestSnippetRepository_Update(t *testing.T) {
	t.Parallel()

	t.Run("writes the edited Snippet and its Fragment", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		stored := insertSnippet(t, repository, testkit.NewSequentialIDs(), testkit.SnippetSpec{Title: "curl"})
		edited := editedSnippet(t, stored, "if x {\n\treturn\n}\n")

		require.NoError(t, repository.Update(t.Context(), edited, stored.UpdatedAt()))

		got, err := repository.Find(t.Context(), stored.ID())
		require.NoError(t, err)
		assert.Equal(t, edited, got)
	})

	t.Run("keeps the Tags the Snippet carries", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		repository := newSnippetRepository(t, database)
		ids := testkit.NewSequentialIDs()
		golang := insertTag(t, newTagRepository(t, database), ids, "go")
		stored := insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang}})

		require.NoError(t, repository.Update(t.Context(), editedSnippet(t, stored, "go"), stored.UpdatedAt()))

		got, err := repository.Find(t.Context(), stored.ID())
		require.NoError(t, err)
		assert.Equal(t, []domain.Tag{golang}, got.Tags())
	})

	t.Run("refuses a save over a Snippet changed since it was loaded", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		stored := insertSnippet(t, repository, testkit.NewSequentialIDs(), testkit.SnippetSpec{Title: "curl"})
		elsewhere := editedSnippet(t, stored, "changed elsewhere")
		require.NoError(t, repository.Update(t.Context(), elsewhere, stored.UpdatedAt()))

		err := repository.Update(t.Context(), editedSnippet(t, stored, "stale"), stored.UpdatedAt())

		require.ErrorIs(t, err, domain.ErrConflict)
		got, findErr := repository.Find(t.Context(), stored.ID())
		require.NoError(t, findErr)
		assert.Equal(t, elsewhere, got)
	})

	t.Run("reports a Snippet that is gone as not found", func(t *testing.T) {
		t.Parallel()

		repository := newSnippetRepository(t, openDatabase(t, newDatabasePath(t)))
		missing := testkit.Snippet(t, testkit.SnippetSpec{})

		err := repository.Update(t.Context(), editedSnippet(t, missing, "gone"), missing.UpdatedAt())

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

	t.Run("returns each Snippet with the Tags it carries", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		repository := newSnippetRepository(t, database)
		tags := newTagRepository(t, database)
		ids := testkit.NewSequentialIDs()
		golang := insertTag(t, tags, ids, "go")
		docker := insertTag(t, tags, ids, "docker")
		tagged := insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, docker}})
		untagged := insertSnippet(t, repository, ids, testkit.SnippetSpec{})

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Snippet{tagged, untagged}, got)
	})

	t.Run("skips a Snippet carrying a corrupt Tag", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newLoggingSnippetRepository(t, path, &logged)
		ids := testkit.NewSequentialIDs()
		broken := insertTag(t, newTagRepository(t, openDatabase(t, path)), ids, "docker")
		healthy := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "healthy"})
		corrupt := insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{broken}})
		execRaw(t, path, "UPDATE tags SET name = 'docker,compose' WHERE id = ?", broken.ID().String())

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []domain.Snippet{healthy}, got)
		assertCorruptRowLogged(t, &logged, "WARN", corrupt.ID())
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
			golang := insertTag(t, newTagRepository(t, openDatabase(t, path)), ids, "go")
			insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "at the Root", Tags: []domain.Tag{golang}})
			filed := insertSnippet(t, repository, ids, testkit.SnippetSpec{
				Title: "filed", FolderID: folderID, Tags: []domain.Tag{golang},
			})

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

func TestSnippetRepository_ListWithTag(t *testing.T) {
	t.Parallel()

	for _, order := range []domain.SortOrder{domain.SortOrderTitle, domain.SortOrderUpdated, domain.SortOrderCreated} {
		t.Run("returns the Snippets carrying the Tag across Folders by "+order.String(), func(t *testing.T) {
			t.Parallel()

			path := newDatabasePath(t)
			database := openDatabase(t, path)
			repository := newSnippetRepository(t, database)
			tags := newTagRepository(t, database)
			ids := testkit.NewSequentialIDs()
			folderID := ids.NewFolderID()
			insertRawFolder(t, path, folderID)
			golang := insertTag(t, tags, ids, "go")
			docker := insertTag(t, tags, ids, "docker")
			atRoot := insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "a", Tags: []domain.Tag{golang}})
			filed := insertSnippet(t, repository, ids, testkit.SnippetSpec{
				Title: "b", FolderID: folderID, Tags: []domain.Tag{golang, docker},
			})
			insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "c", Tags: []domain.Tag{docker}})
			insertSnippet(t, repository, ids, testkit.SnippetSpec{Title: "d"})

			got, err := repository.ListWithTag(t.Context(), golang.ID(), order)

			require.NoError(t, err)
			assert.ElementsMatch(t, []domain.Snippet{atRoot, filed}, got)
		})
	}

	t.Run("orders the Snippets like a Folder listing", func(t *testing.T) {
		t.Parallel()

		earlier := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
		database := openDatabase(t, newDatabasePath(t))
		repository := newSnippetRepository(t, database)
		ids := testkit.NewSequentialIDs()
		golang := insertTag(t, newTagRepository(t, database), ids, "go")
		insertSnippet(t, repository, ids, testkit.SnippetSpec{
			Title: "zebra", CreatedAt: earlier, Tags: []domain.Tag{golang},
		})
		insertSnippet(t, repository, ids, testkit.SnippetSpec{
			Title: "Apple", CreatedAt: earlier.Add(time.Hour), Tags: []domain.Tag{golang},
		})

		byTitle, err := repository.ListWithTag(t.Context(), golang.ID(), domain.SortOrderTitle)
		require.NoError(t, err)

		byCreated, err := repository.ListWithTag(t.Context(), golang.ID(), domain.SortOrderCreated)
		require.NoError(t, err)

		assert.Equal(t, []string{"Apple", "zebra"}, titlesOf(byTitle))
		assert.Equal(t, []string{"Apple", "zebra"}, titlesOf(byCreated))
	})

	t.Run("returns no Snippets for a Tag no Snippet carries", func(t *testing.T) {
		t.Parallel()

		database := openDatabase(t, newDatabasePath(t))
		repository := newSnippetRepository(t, database)
		ids := testkit.NewSequentialIDs()
		unused := insertTag(t, newTagRepository(t, database), ids, "unused")
		insertSnippet(t, repository, ids, testkit.SnippetSpec{})

		got, err := repository.ListWithTag(t.Context(), unused.ID(), domain.SortOrderTitle)

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestSnippetRepository_CountByTag(t *testing.T) {
	t.Parallel()

	database := openDatabase(t, newDatabasePath(t))
	repository := newSnippetRepository(t, database)
	tags := newTagRepository(t, database)
	ids := testkit.NewSequentialIDs()
	golang := insertTag(t, tags, ids, "go")
	docker := insertTag(t, tags, ids, "docker")
	insertTag(t, tags, ids, "unused")
	insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang, docker}})
	insertSnippet(t, repository, ids, testkit.SnippetSpec{Tags: []domain.Tag{golang}})

	got, err := repository.CountByTag(t.Context())

	require.NoError(t, err)
	assert.Equal(t, map[domain.TagID]int{golang.ID(): 2, docker.ID(): 1}, got)
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
