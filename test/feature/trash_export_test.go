//go:build feature

package feature_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/exportfile"
	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const versionOneExport = `{
  "format": "tuisnip-export",
  "version": 1,
  "folders": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000001",
      "parent_id": null,
      "name": "scripts",
      "default_language": "Bash",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z"
    }
  ],
  "tags": [],
  "snippets": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000003",
      "folder_id": "0194c3a0-0000-7000-8000-000000000001",
      "title": "loop",
      "description": "",
      "tag_ids": [],
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z",
      "fragments": [
        {
          "id": "0194c3a0-0000-7000-8000-000000000004",
          "title": "",
          "language": "Bash",
          "content": "ls\n",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-01T09:30:00.123456789Z"
        }
      ]
    }
  ]
}
`

func TestImportingAnExportGivesBackTheSameTrash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	seeded := filepath.Join(dir, "seeded.json")
	require.NoError(t, exportfile.NewWriter().Write(t.Context(), seeded, libraryWithTrash(t)))
	_, first := testapp.Start(t, testapp.RecordingTool)

	counts, err := first.Import.Run(t.Context(), library.ImportInput{Path: seeded})

	require.NoError(t, err)
	assert.Equal(t, library.TrashCounts{Folders: 2, Snippets: 4}, counts.InTrash)
	firstExport := exported(t, first, filepath.Join(dir, "first.json"))
	assert.Equal(t, readFile(t, seeded), string(firstExport))

	_, second := testapp.Start(t, testapp.RecordingTool)
	_, err = second.Import.Run(t.Context(), library.ImportInput{Path: filepath.Join(dir, "first.json")})
	require.NoError(t, err)
	assert.Equal(t, string(firstExport), string(exported(t, second, filepath.Join(dir, "second.json"))))
}

func TestImportOfAVersion1ExportLeavesTheTrashEmpty(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "v1.json")
	require.NoError(t, os.WriteFile(path, []byte(versionOneExport), 0o600))
	_, app := testapp.Start(t, testapp.RecordingTool)

	counts, err := app.Import.Run(t.Context(), library.ImportInput{Path: path})

	require.NoError(t, err)
	assert.Equal(t, library.Counts{
		Folders: 1, Tags: 0, Snippets: 1, Fragments: 1, InTrash: library.TrashCounts{Folders: 0, Snippets: 0},
	}, counts)
	assert.Len(t, testapp.StoredFolders(t, app), 1)
	assert.Len(t, testapp.StoredSnippets(t, app), 1)
}

func TestImportOfARecordDeletedBeforeItWasCreatedNamesItAndWritesNothing(t *testing.T) {
	t.Parallel()

	_, source := testapp.Start(t, testapp.RecordingTool)
	note := testapp.SeedSnippet(t, source, snippet.CreateInput{Title: "note", Fragments: plainFragment("remember")})
	dir := t.TempDir()
	path := filepath.Join(dir, "source.json")
	edited := strings.Replace(
		string(exported(t, source, path)), `"deleted_at": null`, `"deleted_at": "2000-01-01T00:00:00.000000000Z"`, 1,
	)
	require.NoError(t, os.WriteFile(path, []byte(edited), 0o600))
	_, target := testapp.Start(t, testapp.RecordingTool)
	before := exported(t, target, filepath.Join(dir, "before.json"))

	_, err := target.Import.Run(t.Context(), library.ImportInput{Path: path})

	require.ErrorIs(t, err, domain.ErrDeletedBeforeCreate)
	require.ErrorContains(t, err, "snippet "+note.ID().String()+": ")
	assert.Equal(t, string(before), string(exported(t, target, filepath.Join(dir, "after.json"))))
}

func TestImportIntoALibraryHoldingOnlyTrashItemsIsRefused(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	onlyTrash := filepath.Join(dir, "trash.json")
	require.NoError(t, exportfile.NewWriter().Write(t.Context(), onlyTrash, libraryWithTrash(t)))
	_, target := testapp.Start(t, testapp.RecordingTool)
	_, err := target.Import.Run(t.Context(), library.ImportInput{Path: onlyTrash})
	require.NoError(t, err)

	_, err = target.Import.Run(t.Context(), library.ImportInput{Path: onlyTrash})

	assert.ErrorIs(t, err, domain.ErrLibraryNotEmpty)
}

func libraryWithTrash(t *testing.T) domain.Library {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	created := time.Date(2026, time.March, 1, 9, 30, 0, 123456789, time.UTC)
	deleted := created.Add(48 * time.Hour)
	trashedFolder := testkit.Folder(t, testkit.FolderSpec{ID: ids.NewFolderID(), Name: "docker", CreatedAt: created})
	subfolder := testkit.Folder(t, testkit.FolderSpec{
		ID: ids.NewFolderID(), Name: "compose", ParentID: trashedFolder.ID(), CreatedAt: created,
	})
	trashedSnippet := testkit.Snippet(t, testkit.SnippetSpec{
		ID: ids.NewSnippetID(), Title: "prune", Fragments: oneFragment(ids, created), CreatedAt: created,
	})
	insideTrashedFolder := testkit.Snippet(t, testkit.SnippetSpec{
		ID: ids.NewSnippetID(), Title: "up", FolderID: subfolder.ID(), Fragments: oneFragment(ids, created),
		CreatedAt: created,
	})
	separatelyTrashed := testkit.Snippet(t, testkit.SnippetSpec{
		ID: ids.NewSnippetID(), Title: "logs", FolderID: trashedFolder.ID(), Fragments: oneFragment(ids, created),
		CreatedAt: created,
	})
	detached := testkit.Snippet(t, testkit.SnippetSpec{
		ID: ids.NewSnippetID(), Title: "build", Fragments: oneFragment(ids, created), CreatedAt: created,
	})

	return domain.Library{
		Folders:  []domain.Folder{trashedFolder, subfolder},
		Tags:     nil,
		Snippets: []domain.Snippet{trashedSnippet, insideTrashedFolder, separatelyTrashed, detached},
		Trash: domain.Trash{
			Folders: map[domain.FolderID]domain.TrashMark{
				trashedFolder.ID(): testkit.TrashMark(
					t,
					testkit.TrashMarkSpec{Item: trashedFolder, DeletedAt: deleted},
				),
			},
			Snippets: map[domain.SnippetID]domain.TrashMark{
				trashedSnippet.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
					Item: trashedSnippet, DeletedAt: deleted.Add(time.Minute),
				}),
				separatelyTrashed.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
					Item: separatelyTrashed, DeletedAt: deleted.Add(-time.Minute),
				}),
				detached.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
					Item: detached, DeletedAt: deleted, FormerPath: []string{"scripts", "ci/cd"},
				}),
			},
		},
	}
}

func oneFragment(ids *testkit.SequentialIDs, created time.Time) []testkit.FragmentSpec {
	return []testkit.FragmentSpec{{ID: ids.NewFragmentID(), Content: "docker system prune", CreatedAt: created}}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}
