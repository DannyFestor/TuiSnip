package exportfile_test

import (
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/exportfile"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const (
	folderID   = "0194c3a0-0000-7000-8000-000000000001"
	tagID      = "0194c3a0-0000-7000-8000-000000000002"
	snippetID  = "0194c3a0-0000-7000-8000-000000000003"
	fragmentID = "0194c3a0-0000-7000-8000-000000000004"
	missingID  = "0194c3a0-0000-7000-8000-0000000000ff"
	createdAt  = "2026-03-01T09:30:00.123456789Z"
	updatedAt  = "2026-03-02T09:00:00.000000000Z"
	deletedAt  = "2026-03-03T09:00:00.000000005Z"
)

type record = map[string]any

func TestReader_Read(t *testing.T) {
	t.Parallel()

	t.Run("reads back the Library the Writer wrote, IDs and dates included", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		created := time.Date(2026, time.March, 1, 9, 30, 0, 123456789, time.UTC)
		updated := time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)
		parent := testkit.Folder(
			t,
			testkit.FolderSpec{ID: ids.NewFolderID(), DefaultLanguage: "Bash", CreatedAt: created},
		)
		child := testkit.Folder(t, testkit.FolderSpec{
			ID: ids.NewFolderID(), ParentID: parent.ID(), CreatedAt: created, UpdatedAt: updated,
		})
		tag := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Shell", CreatedAt: created})
		filed := testkit.Snippet(t, testkit.SnippetSpec{
			ID:          ids.NewSnippetID(),
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
		atRoot := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: updated})
		detached := testkit.Snippet(t, testkit.SnippetSpec{ID: ids.NewSnippetID(), CreatedAt: updated})
		written := domain.Library{
			Folders:  []domain.Folder{parent, child},
			Tags:     []domain.Tag{tag},
			Snippets: []domain.Snippet{filed, atRoot, detached},
			Trash: domain.Trash{
				Folders: map[domain.FolderID]domain.TrashMark{
					child.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{Item: child, DeletedAt: updated}),
				},
				Snippets: map[domain.SnippetID]domain.TrashMark{
					filed.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{Item: filed}),
					detached.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
						Item: detached, FormerPath: []string{"scripts", "a/b"},
					}),
				},
			},
		}
		path := filepath.Join(t.TempDir(), exportName)
		require.NoError(t, exportfile.NewWriter().Write(t.Context(), path, written))

		got, err := exportfile.NewReader().Read(t.Context(), path)

		require.NoError(t, err)
		assert.Equal(t, written, got)
	})

	t.Run("reads a Snippet at the Root carrying a Tag", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, document(t, []record{}, []record{tagRecord(nil)}, []record{snippetRecord(nil)}))

		got, err := exportfile.NewReader().Read(t.Context(), path)

		require.NoError(t, err)
		require.Len(t, got.Snippets, 1)
		assert.True(t, got.Snippets[0].AtRoot())
		assert.Equal(t, got.Tags, got.Snippets[0].Tags())
		assert.Equal(t, "loop", got.Snippets[0].Title().String())
	})

	t.Run("reads a version 1 Export as a Library with an empty Trash", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, versionOneDocument(
			t,
			[]record{versionOne(folderRecord(nil))},
			[]record{tagRecord(nil)},
			[]record{versionOne(snippetRecord(record{"folder_id": folderID}))},
		))

		got, err := exportfile.NewReader().Read(t.Context(), path)

		require.NoError(t, err)
		require.Len(t, got.Folders, 1)
		require.Len(t, got.Snippets, 1)
		assert.Equal(t, got.Folders[0].ID(), got.Snippets[0].FolderID())
		assert.Empty(t, got.InTrash().Folders)
		assert.Empty(t, got.InTrash().Snippets)
	})

	t.Run("reads a Trash item kept in a Folder the Trash holds", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, document(
			t,
			[]record{folderRecord(record{"deleted_at": deletedAt})},
			[]record{tagRecord(nil)},
			[]record{snippetRecord(record{"folder_id": folderID, "deleted_at": deletedAt})},
		))

		got, err := exportfile.NewReader().Read(t.Context(), path)

		require.NoError(t, err)
		require.Len(t, got.Trash.Folders, 1)
		require.Len(t, got.Trash.Snippets, 1)
		assert.Equal(t, deletedAt, got.Trash.Snippets[got.Snippets[0].ID()].DeletedAt().Format(time.RFC3339Nano))
	})

	refusals := []struct {
		name        string
		document    string
		wantErr     error
		wantMessage string
	}{
		{
			name:        "refuses JSON that isn't a TuiSnip Export",
			document:    `{"format": "snippetslab", "version": 1}`,
			wantErr:     exportfile.ErrNotAnExport,
			wantMessage: "",
		},
		{
			name:        "refuses a newer version and asks for an upgrade",
			document:    `{"format": "tuisnip-export", "version": 3, "added_later": []}`,
			wantErr:     exportfile.ErrNewerVersion,
			wantMessage: "upgrade TuiSnip",
		},
		{
			name: "refuses a version 1 Export carrying a field only version 2 has",
			document: versionOneDocument(
				t,
				[]record{overridden(versionOne(folderRecord(nil)), record{"deleted_at": nil})},
				[]record{},
				[]record{},
			),
			wantErr:     nil,
			wantMessage: `unknown field "deleted_at"`,
		},
		{
			name: "refuses a Snippet deleted before it was created and names it",
			document: document(t, []record{}, []record{tagRecord(nil)}, []record{
				snippetRecord(record{"deleted_at": "2026-02-28T09:00:00.000000000Z"}),
			}),
			wantErr:     domain.ErrDeletedBeforeCreate,
			wantMessage: "snippet " + snippetID + ": ",
		},
		{
			name: "refuses a Folder deleted before it was created and names it",
			document: document(
				t,
				[]record{folderRecord(record{"deleted_at": "2026-02-28T09:00:00.000000000Z"})},
				[]record{},
				[]record{},
			),
			wantErr:     domain.ErrDeletedBeforeCreate,
			wantMessage: "folder " + folderID + ": ",
		},
		{
			name: "refuses a deletion date it can't parse and names the Folder",
			document: document(
				t, []record{folderRecord(record{"deleted_at": "yesterday"})}, []record{}, []record{},
			),
			wantErr:     nil,
			wantMessage: "folder " + folderID + ": ",
		},
		{
			name: "refuses a former path on a Snippet outside the Trash and names it",
			document: document(t, []record{}, []record{tagRecord(nil)}, []record{
				snippetRecord(record{"former_folder_path": []string{"scripts"}}),
			}),
			wantErr:     domain.ErrFormerPathOutsideTrash,
			wantMessage: "snippet " + snippetID + ": ",
		},
		{
			name: "refuses a former path on a Trash item still in its Folder and names it",
			document: document(t, []record{folderRecord(nil)}, []record{tagRecord(nil)}, []record{
				snippetRecord(record{
					"folder_id": folderID, "deleted_at": deletedAt, "former_folder_path": []string{"scripts"},
				}),
			}),
			wantErr:     domain.ErrFormerPathNotDetached,
			wantMessage: "snippet " + snippetID + ": ",
		},
		{
			name: "refuses a blank Folder name in a former path and names the Folder",
			document: document(t, []record{
				folderRecord(record{"deleted_at": deletedAt, "former_parent_path": []string{" "}}),
			}, []record{}, []record{}),
			wantErr:     value.ErrBlankFolderName,
			wantMessage: "folder " + folderID + ": ",
		},
		{
			name:        "refuses a version it never wrote",
			document:    `{"format": "tuisnip-export", "version": 0}`,
			wantErr:     exportfile.ErrUnknownVersion,
			wantMessage: "",
		},
		{
			name:        "refuses a field the format lacks",
			document:    document(t, []record{}, []record{tagRecord(record{"colour": "red"})}, []record{}),
			wantErr:     nil,
			wantMessage: `unknown field "colour"`,
		},
		{
			name: "refuses a Snippet title over 200 characters and names the Snippet",
			document: document(t, []record{}, []record{tagRecord(nil)}, []record{
				snippetRecord(record{"title": strings.Repeat("a", 201)}),
			}),
			wantErr:     value.ErrTitleTooLong,
			wantMessage: "snippet " + snippetID + ": ",
		},
		{
			name: "refuses an unknown Language and names the Snippet and its Fragment",
			document: document(t, []record{}, []record{tagRecord(nil)}, []record{
				snippetRecord(record{"fragments": []record{fragmentRecord(record{"language": "Klingon"})}}),
			}),
			wantErr:     value.ErrUnknownLanguage,
			wantMessage: "snippet " + snippetID + ": fragment " + fragmentID + ": ",
		},
		{
			name: "refuses a Folder updated before it was created and names it",
			document: document(
				t,
				[]record{
					folderRecord(record{"created_at": updatedAt, "updated_at": createdAt}),
				},
				[]record{},
				[]record{},
			),
			wantErr:     domain.ErrUpdatedBeforeCreate,
			wantMessage: "folder " + folderID + ": ",
		},
		{
			name:        "refuses a malformed Tag ID and names it",
			document:    document(t, []record{}, []record{tagRecord(record{"id": "tag-1"})}, []record{}),
			wantErr:     nil,
			wantMessage: "tag tag-1: ",
		},
		{
			name:        "refuses a Snippet carrying a Tag the file lacks and names both",
			document:    document(t, []record{}, []record{}, []record{snippetRecord(nil)}),
			wantErr:     domain.ErrMissingTag,
			wantMessage: "snippet " + snippetID + ": tag " + tagID + ": ",
		},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := exportfile.NewReader().Read(t.Context(), exportFile(t, tt.document))

			require.ErrorContains(t, err, "exportfile.Reader.Read: ")
			require.ErrorContains(t, err, tt.wantMessage)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}

	t.Run("names every bad record at once", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, document(
			t,
			[]record{folderRecord(record{"name": " "})},
			[]record{tagRecord(record{"name": "a,b"})},
			[]record{},
		))

		_, err := exportfile.NewReader().Read(t.Context(), path)

		require.ErrorIs(t, err, value.ErrBlankFolderName)
		require.ErrorIs(t, err, value.ErrTagNameHasComma)
		require.ErrorContains(t, err, "folder "+folderID+": ")
		assert.ErrorContains(t, err, "tag "+tagID+": ")
	})

	t.Run("fails on a missing file", func(t *testing.T) {
		t.Parallel()

		_, err := exportfile.NewReader().Read(t.Context(), filepath.Join(t.TempDir(), exportName))

		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func exportFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), exportName)
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

	return path
}

func document(t *testing.T, folders, tags, snippets []record) string {
	t.Helper()

	return encoded(t, documentRecord(folders, tags, snippets))
}

func versionOneDocument(t *testing.T, folders, tags, snippets []record) string {
	t.Helper()

	return encoded(t, overridden(documentRecord(folders, tags, snippets), record{"version": 1}))
}

func documentRecord(folders, tags, snippets []record) record {
	return record{
		"format":   "tuisnip-export",
		"version":  2,
		"folders":  folders,
		"tags":     tags,
		"snippets": snippets,
	}
}

func encoded(t *testing.T, document record) string {
	t.Helper()

	data, err := json.Marshal(document)
	require.NoError(t, err)

	return string(data)
}

func versionOne(versionTwo record) record {
	withoutTrash := maps.Clone(versionTwo)
	for _, field := range []string{"deleted_at", "former_parent_path", "former_folder_path"} {
		delete(withoutTrash, field)
	}

	return withoutTrash
}

func folderRecord(overrides record) record {
	return overridden(record{
		"id":                 folderID,
		"parent_id":          nil,
		"name":               "scripts",
		"default_language":   "Bash",
		"created_at":         createdAt,
		"updated_at":         updatedAt,
		"deleted_at":         nil,
		"former_parent_path": []string{},
	}, overrides)
}

func tagRecord(overrides record) record {
	return overridden(record{"id": tagID, "name": "Shell", "created_at": createdAt, "updated_at": createdAt}, overrides)
}

func snippetRecord(overrides record) record {
	return overridden(record{
		"id":                 snippetID,
		"folder_id":          nil,
		"title":              "loop",
		"description":        "",
		"tag_ids":            []string{tagID},
		"created_at":         createdAt,
		"updated_at":         updatedAt,
		"fragments":          []record{fragmentRecord(nil)},
		"deleted_at":         nil,
		"former_folder_path": []string{},
	}, overrides)
}

func fragmentRecord(overrides record) record {
	return overridden(record{
		"id":         fragmentID,
		"title":      "",
		"language":   "Bash",
		"content":    "for f in *; do echo \"$f\"; done\n",
		"created_at": createdAt,
		"updated_at": updatedAt,
	}, overrides)
}

func overridden(defaults, overrides record) record {
	maps.Copy(defaults, overrides)

	return defaults
}
