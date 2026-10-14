//go:build feature

package feature_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/exportfile"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
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
	assert.Equal(t, library.Counts{Folders: 1, Tags: 2, Snippets: 1, Fragments: 1}, counts)
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

func TestImportingAnExportIntoAnEmptyLibraryExportsTheSameFile(t *testing.T) {
	t.Parallel()

	_, source := testapp.Start(t, testapp.RecordingTool)
	subfolder := testapp.SeedFolder(t, source, folder.CreateInput{Name: "go", ParentID: domain.FolderID{}})
	parent := testapp.SeedFolder(t, source, folder.CreateInput{Name: "scripts", ParentID: domain.FolderID{}})
	_, err := source.MoveFolder.Run(t.Context(), folder.MoveInput{FolderID: subfolder.ID(), ParentID: parent.ID()})
	require.NoError(t, err)

	shell := testapp.SeedTag(t, source, tag.CreateInput{Name: "shell"})
	testapp.SeedTag(t, source, tag.CreateInput{Name: "unused"})
	testapp.SeedSnippet(t, source, snippet.CreateInput{
		Title:     "loop",
		FolderID:  subfolder.ID(),
		Fragments: append(fragmentIn("Bash", "for f in *; do echo \"$f\"; done\n"), plainFragment("notes")...),
		Tags:      []domain.Tag{shell},
	})
	testapp.SeedSnippet(t, source, snippet.CreateInput{Title: "note", Fragments: plainFragment("a < b & c")})
	dir := t.TempDir()
	first := exported(t, source, filepath.Join(dir, "first.json"))
	_, target := testapp.Start(t, testapp.RecordingTool)

	counts, err := target.Import.Run(t.Context(), library.ImportInput{Path: filepath.Join(dir, "first.json")})

	require.NoError(t, err)
	assert.Equal(t, library.Counts{Folders: 2, Tags: 2, Snippets: 2, Fragments: 3}, counts)
	assert.Equal(t, string(first), string(exported(t, target, filepath.Join(dir, "second.json"))))
}

func TestImportIntoALibraryHoldingATagIsRefusedAndChangesNothing(t *testing.T) {
	t.Parallel()

	_, source := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedSnippet(t, source, snippet.CreateInput{Title: "note", Fragments: plainFragment("remember")})
	dir := t.TempDir()
	exported(t, source, filepath.Join(dir, "source.json"))
	_, target := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedTag(t, target, tag.CreateInput{Name: "shell"})
	before := exported(t, target, filepath.Join(dir, "before.json"))

	_, err := target.Import.Run(t.Context(), library.ImportInput{Path: filepath.Join(dir, "source.json")})

	require.ErrorIs(t, err, domain.ErrLibraryNotEmpty)
	assert.Equal(t, string(before), string(exported(t, target, filepath.Join(dir, "after.json"))))
}

func TestImportOfATitleOver200CharactersNamesTheSnippetAndWritesNothing(t *testing.T) {
	t.Parallel()

	_, source := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedTag(t, source, tag.CreateInput{Name: "shell"})
	long := testapp.SeedSnippet(t, source, snippet.CreateInput{
		Title: strings.Repeat("a", 200), Fragments: plainFragment("remember"),
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "source.json")
	edited := strings.Replace(string(exported(t, source, path)), long.Title().String(), strings.Repeat("a", 201), 1)
	require.NoError(t, os.WriteFile(path, []byte(edited), 0o600))
	_, target := testapp.Start(t, testapp.RecordingTool)
	before := exported(t, target, filepath.Join(dir, "before.json"))

	_, err := target.Import.Run(t.Context(), library.ImportInput{Path: path})

	require.ErrorIs(t, err, value.ErrTitleTooLong)
	require.ErrorContains(t, err, "snippet "+long.ID().String()+": ")
	assert.Equal(t, string(before), string(exported(t, target, filepath.Join(dir, "after.json"))))
}

func TestImportOfANewerExportVersionIsRefused(t *testing.T) {
	t.Parallel()

	_, source := testapp.Start(t, testapp.RecordingTool)
	path := filepath.Join(t.TempDir(), "source.json")
	edited := strings.Replace(string(exported(t, source, path)), `"version": 1`, `"version": 2`, 1)
	require.NoError(t, os.WriteFile(path, []byte(edited), 0o600))
	_, target := testapp.Start(t, testapp.RecordingTool)

	_, err := target.Import.Run(t.Context(), library.ImportInput{Path: path})

	assert.ErrorIs(t, err, exportfile.ErrNewerVersion)
}

func exported(t *testing.T, app *bootstrap.App, path string) []byte {
	t.Helper()

	_, err := app.Export.Run(t.Context(), library.ExportInput{Path: path})
	require.NoError(t, err)

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return data
}
