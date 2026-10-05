package sqlite_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
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
