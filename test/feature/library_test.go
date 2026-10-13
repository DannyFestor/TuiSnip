//go:build feature

package feature_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

type exportedTags struct {
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
}

func TestExportingAnUnchangedLibraryTwiceWritesIdenticalFiles(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	scripts := testapp.SeedFolder(t, app, folder.CreateInput{Name: "scripts", ParentID: domain.FolderID{}})
	shell := testapp.SeedTag(t, app, tag.CreateInput{Name: "shell"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{
		Title:     "loop",
		FolderID:  scripts.ID(),
		Fragments: fragmentIn("Bash", "for f in *; do echo \"$f\"; done\n"),
		Tags:      []domain.Tag{shell},
	})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "note", Fragments: plainFragment("remember")})
	dir := t.TempDir()

	first := exported(t, app, filepath.Join(dir, "first.json"))
	second := exported(t, app, filepath.Join(dir, "second.json"))

	assert.Equal(t, first, second)
}

func TestExportReportsTheCountsWritten(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedFolder(t, app, folder.CreateInput{Name: "scripts", ParentID: domain.FolderID{}})
	testapp.SeedTag(t, app, tag.CreateInput{Name: "shell"})
	testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "note", Fragments: plainFragment("remember")})

	counts, err := app.Export.Run(t.Context(), library.ExportInput{Path: filepath.Join(t.TempDir(), "out.json")})

	require.NoError(t, err)
	assert.Equal(t, library.ExportCounts{Folders: 1, Tags: 2, Snippets: 1}, counts)
}

func TestExportIncludesATagNoSnippetCarries(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedTag(t, app, tag.CreateInput{Name: "unused"})
	path := filepath.Join(t.TempDir(), "out.json")

	var written exportedTags
	require.NoError(t, json.Unmarshal(exported(t, app, path), &written))

	require.Len(t, written.Tags, 1)
	assert.Equal(t, "unused", written.Tags[0].Name)
}

func exported(t *testing.T, app *bootstrap.App, path string) []byte {
	t.Helper()

	_, err := app.Export.Run(t.Context(), library.ExportInput{Path: path})
	require.NoError(t, err)

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return data
}
