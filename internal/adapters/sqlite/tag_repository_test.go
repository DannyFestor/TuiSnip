package sqlite_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestTagRepository_Insert(t *testing.T) {
	t.Parallel()

	t.Run("refuses a name that differs only in case, keeping the first spelling", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		first := insertTag(t, repository, ids, "Go")

		err := repository.Insert(t.Context(), testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}))

		require.ErrorIs(t, err, domain.ErrTagNameTaken)

		got, err := repository.List(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []domain.Tag{first}, got)
	})

	t.Run("refuses a name that differs only in case beyond ASCII", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		insertTag(t, repository, ids, "Ärger")

		err := repository.Insert(t.Context(), testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "ärger"}))

		require.ErrorIs(t, err, domain.ErrTagNameTaken)
	})

	t.Run("passes any other failure on unchanged", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		existing := insertTag(t, repository, ids, "go")

		err := repository.Insert(t.Context(), testkit.Tag(t, testkit.TagSpec{ID: existing.ID(), Name: "docker"}))

		require.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrTagNameTaken)
	})
}

func TestTagRepository_Find(t *testing.T) {
	t.Parallel()

	t.Run("returns the Tag", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))
		stored := insertTag(t, repository, testkit.NewSequentialIDs(), "go")

		got, err := repository.Find(t.Context(), stored.ID())

		require.NoError(t, err)
		assert.Equal(t, stored, got)
	})

	t.Run("reports a missing Tag as not found", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))

		_, err := repository.Find(t.Context(), testkit.NewSequentialIDs().NewTagID())

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.TagRepository.Find")
	})

	t.Run("reports and logs a corrupt Tag", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newLoggingTagRepository(t, path, &logged)
		corrupt := insertTag(t, repository, testkit.NewSequentialIDs(), "docker")
		execRaw(t, path, "UPDATE tags SET name = 'docker,compose' WHERE id = ?", corrupt.ID().String())

		_, err := repository.Find(t.Context(), corrupt.ID())

		require.ErrorIs(t, err, domain.ErrCorruptRecord)
		assert.Contains(t, logged.String(), `"level":"ERROR"`)
		assert.Contains(t, logged.String(), `"tag_id":"`+corrupt.ID().String()+`"`)
	})
}

func TestTagRepository_Rename(t *testing.T) {
	t.Parallel()

	t.Run("stores the new name and time", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		stored := insertTag(t, fixture.tags, fixture.ids, "go")
		renamed := renamedTag(t, stored, "golang")

		survivor, err := fixture.tags.Rename(t.Context(), renamed)

		require.NoError(t, err)
		assert.Equal(t, renamed, survivor)
		assert.Equal(t, []domain.Tag{renamed}, listedTags(t, fixture.tags))
	})

	t.Run("changes only the case of its own name", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		stored := insertTag(t, fixture.tags, fixture.ids, "go")
		renamed := renamedTag(t, stored, "Go")

		survivor, err := fixture.tags.Rename(t.Context(), renamed)

		require.NoError(t, err)
		assert.Equal(t, renamed, survivor)
		assert.Equal(t, []domain.Tag{renamed}, listedTags(t, fixture.tags))
	})

	t.Run("merges onto the Tag that has the name, keeping its ID and taking the typed spelling", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		existing := insertTag(t, fixture.tags, fixture.ids, "go")
		merged := insertTag(t, fixture.tags, fixture.ids, "golang")
		renamed := renamedTag(t, merged, "Go")

		survivor, err := fixture.tags.Rename(t.Context(), renamed)

		require.NoError(t, err)
		assert.Equal(t, existing.ID(), survivor.ID())
		assert.Equal(t, "Go", survivor.Name().String())
		assert.Equal(t, existing.CreatedAt(), survivor.CreatedAt())
		assert.Equal(t, renamed.UpdatedAt(), survivor.UpdatedAt())
		assert.Equal(t, []domain.Tag{survivor}, listedTags(t, fixture.tags))
	})

	t.Run("leaves each Snippet of a merge carrying the surviving Tag exactly once", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		existing := insertTag(t, fixture.tags, fixture.ids, "go")
		merged := insertTag(t, fixture.tags, fixture.ids, "golang")
		both := fixture.insertSnippetWith(t, existing, merged)
		onlyMerged := fixture.insertSnippetWith(t, merged)
		onlyExisting := fixture.insertSnippetWith(t, existing)

		survivor, err := fixture.tags.Rename(t.Context(), renamedTag(t, merged, "go"))

		require.NoError(t, err)

		for _, carrier := range []domain.Snippet{both, onlyMerged, onlyExisting} {
			assert.Equal(t, []domain.Tag{survivor}, fixture.tagsOf(t, carrier))
		}
	})

	t.Run("reports a missing Tag as not found", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		missing := testkit.Tag(t, testkit.TagSpec{ID: fixture.ids.NewTagID(), Name: "go"})

		_, err := fixture.tags.Rename(t.Context(), missing)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.TagRepository.Rename")
	})

	t.Run("reports a missing Tag as not found when the name belongs to another", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		existing := insertTag(t, fixture.tags, fixture.ids, "go")
		missing := testkit.Tag(t, testkit.TagSpec{ID: fixture.ids.NewTagID(), Name: "Go"})

		_, err := fixture.tags.Rename(t.Context(), missing)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Equal(t, []domain.Tag{existing}, listedTags(t, fixture.tags))
	})
}

func TestTagRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("removes the Tag from every Snippet and keeps the Snippets", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		golang := insertTag(t, fixture.tags, fixture.ids, "go")
		docker := insertTag(t, fixture.tags, fixture.ids, "docker")
		carrier := fixture.insertSnippetWith(t, golang, docker)

		require.NoError(t, fixture.tags.Delete(t.Context(), golang.ID()))

		assert.Equal(t, []domain.Tag{docker}, listedTags(t, fixture.tags))
		assert.Equal(t, []domain.Tag{docker}, fixture.tagsOf(t, carrier))
	})

	t.Run("reports a missing Tag as not found", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)

		err := fixture.tags.Delete(t.Context(), fixture.ids.NewTagID())

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "sqlite.TagRepository.Delete")
	})
}

func TestTagRepository_CountSnippets(t *testing.T) {
	t.Parallel()

	t.Run("counts the Snippets carrying the Tag", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		golang := insertTag(t, fixture.tags, fixture.ids, "go")
		docker := insertTag(t, fixture.tags, fixture.ids, "docker")
		fixture.insertSnippetWith(t, golang, docker)
		fixture.insertSnippetWith(t, golang)
		fixture.insertSnippetWith(t, docker)

		count, err := fixture.tags.CountSnippets(t.Context(), golang.ID())

		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("counts no Snippets for a Tag none carries", func(t *testing.T) {
		t.Parallel()

		fixture := newTagFixture(t)
		unused := insertTag(t, fixture.tags, fixture.ids, "unused")

		count, err := fixture.tags.CountSnippets(t.Context(), unused.ID())

		require.NoError(t, err)
		assert.Zero(t, count)
	})
}

type tagFixture struct {
	tags     *sqlite.TagRepository
	snippets *sqlite.SnippetRepository
	ids      *testkit.SequentialIDs
}

func newTagFixture(t *testing.T) tagFixture {
	t.Helper()

	database := openDatabase(t, newDatabasePath(t))

	return tagFixture{
		tags:     newTagRepository(t, database),
		snippets: newSnippetRepository(t, database),
		ids:      testkit.NewSequentialIDs(),
	}
}

func (f tagFixture) insertSnippetWith(t *testing.T, tags ...domain.Tag) domain.Snippet {
	t.Helper()

	return insertSnippet(t, f.snippets, f.ids, testkit.SnippetSpec{Tags: tags})
}

func (f tagFixture) tagsOf(t *testing.T, carrier domain.Snippet) []domain.Tag {
	t.Helper()

	stored, err := f.snippets.Find(t.Context(), carrier.ID())
	require.NoError(t, err)

	return stored.Tags()
}

func renamedTag(t *testing.T, stored domain.Tag, name string) domain.Tag {
	t.Helper()

	tagName, err := value.NewTagName(name)
	require.NoError(t, err)

	renamed, err := stored.Rename(tagName, stored.UpdatedAt().Add(time.Minute))
	require.NoError(t, err)

	return renamed
}

func listedTags(t *testing.T, repository *sqlite.TagRepository) []domain.Tag {
	t.Helper()

	tags, err := repository.List(t.Context())
	require.NoError(t, err)

	return tags
}

func TestTagRepository_List(t *testing.T) {
	t.Parallel()

	t.Run("returns every inserted Tag", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))
		ids := testkit.NewSequentialIDs()
		golang := insertTag(t, repository, ids, "go")
		docker := insertTag(t, repository, ids, "docker")

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(t, []domain.Tag{golang, docker}, got)
	})

	t.Run("returns no Tags from an empty database", func(t *testing.T) {
		t.Parallel()

		repository := newTagRepository(t, openDatabase(t, newDatabasePath(t)))

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("skips a Tag whose name has a comma", func(t *testing.T) {
		t.Parallel()

		var logged bytes.Buffer

		path := newDatabasePath(t)
		repository := newLoggingTagRepository(t, path, &logged)
		ids := testkit.NewSequentialIDs()
		healthy := insertTag(t, repository, ids, "go")
		corrupt := insertTag(t, repository, ids, "docker")
		execRaw(t, path, "UPDATE tags SET name = 'docker,compose' WHERE id = ?", corrupt.ID().String())

		got, err := repository.List(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []domain.Tag{healthy}, got)
		assert.Contains(t, logged.String(), `"level":"WARN"`)
		assert.Contains(t, logged.String(), `"tag_id":"`+corrupt.ID().String()+`"`)
		assert.Contains(t, logged.String(), `"error":"`+domain.ErrCorruptRecord.Error())
	})
}
