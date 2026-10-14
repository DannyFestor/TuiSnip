package exportfile_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/exportfile"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	t.Run("writes the Library as a TuiSnip Export", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		tokyo := time.FixedZone("JST", 9*60*60)
		created := time.Date(2026, time.March, 1, 18, 30, 0, 123456789, tokyo)
		updated := time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)
		parent := testkit.Folder(t, testkit.FolderSpec{
			ID: ids.NewFolderID(), Name: "scripts", DefaultLanguage: "Bash", CreatedAt: created,
		})
		child := testkit.Folder(t, testkit.FolderSpec{
			ID: ids.NewFolderID(), Name: "go", ParentID: parent.ID(), DefaultLanguage: "Go",
			CreatedAt: created, UpdatedAt: updated,
		})
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Shell", CreatedAt: created})
		filed := testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
			Title:       "compare",
			Description: "less <than> & more",
			FolderID:    child.ID(),
			Fragments: []testkit.FragmentSpec{
				{ID: ids.NewFragmentID(), Title: "main.go", Language: "Go", Content: "a < b\n", CreatedAt: created},
				{ID: ids.NewFragmentID(), Content: "notes", CreatedAt: created, UpdatedAt: updated},
			},
			Tags:      []domain.Tag{tag},
			CreatedAt: created,
			UpdatedAt: updated,
		})
		atRoot := testkit.Snippet(t, testkit.SnippetSpec{
			ID:        ids.NewSnippetID(),
			Title:     "loose",
			Fragments: []testkit.FragmentSpec{{ID: ids.NewFragmentID(), Content: "x", CreatedAt: updated}},
			CreatedAt: updated,
		})
		path := filepath.Join(t.TempDir(), exportName)

		err := exportfile.NewWriter().Write(t.Context(), path, domain.Library{
			Folders:  []domain.Folder{parent, child},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{filed, atRoot},
		})

		require.NoError(t, err)
		assert.Equal(t, `{
  "format": "tuisnip-export",
  "version": 1,
  "folders": [
    {
      "id": "`+parent.ID().String()+`",
      "parent_id": null,
      "name": "scripts",
      "default_language": "Bash",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z"
    },
    {
      "id": "`+child.ID().String()+`",
      "parent_id": "`+parent.ID().String()+`",
      "name": "go",
      "default_language": "Go",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-02T09:00:00.000000000Z"
    }
  ],
  "tags": [
    {
      "id": "`+tag.ID().String()+`",
      "name": "Shell",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z"
    }
  ],
  "snippets": [
    {
      "id": "`+filed.ID().String()+`",
      "folder_id": "`+child.ID().String()+`",
      "title": "compare",
      "description": "less <than> & more",
      "tag_ids": [
        "`+tag.ID().String()+`"
      ],
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-02T09:00:00.000000000Z",
      "fragments": [
        {
          "id": "`+filed.Fragments()[0].ID().String()+`",
          "title": "main.go",
          "language": "Go",
          "content": "a < b\n",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-01T09:30:00.123456789Z"
        },
        {
          "id": "`+filed.Fragments()[1].ID().String()+`",
          "title": "",
          "language": "plaintext",
          "content": "notes",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-02T09:00:00.000000000Z"
        }
      ]
    },
    {
      "id": "`+atRoot.ID().String()+`",
      "folder_id": null,
      "title": "loose",
      "description": "",
      "tag_ids": [],
      "created_at": "2026-03-02T09:00:00.000000000Z",
      "updated_at": "2026-03-02T09:00:00.000000000Z",
      "fragments": [
        {
          "id": "`+atRoot.FirstFragment().ID().String()+`",
          "title": "",
          "language": "plaintext",
          "content": "x",
          "created_at": "2026-03-02T09:00:00.000000000Z",
          "updated_at": "2026-03-02T09:00:00.000000000Z"
        }
      ]
    }
  ]
}
`, read(t, path))
	})

	t.Run("writes empty lists for an empty Library", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), exportName)

		require.NoError(t, exportfile.NewWriter().Write(t.Context(), path, domain.Library{}))

		//nolint:testifylint // encoded-compare: the exact bytes are the format, not only the JSON they decode to
		assert.Equal(t, `{
  "format": "tuisnip-export",
  "version": 1,
  "folders": [],
  "tags": [],
  "snippets": []
}
`, read(t, path))
	})

	t.Run("writes a file only its owner can read", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), exportName)

		require.NoError(t, exportfile.NewWriter().Write(t.Context(), path, domain.Library{}))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("writes nothing once the context is cancelled", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), exportName)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := exportfile.NewWriter().Write(ctx, path, domain.Library{})

		require.ErrorIs(t, err, context.Canceled)
		require.ErrorContains(t, err, "exportfile.Writer.Write: ")

		_, statErr := os.Stat(path)
		assert.ErrorIs(t, statErr, fs.ErrNotExist)
	})
}

func read(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}
