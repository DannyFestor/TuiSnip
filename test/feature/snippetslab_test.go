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

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const snippetsLabFixture = "testdata/snippetslab-export.json"

type fragmentView struct {
	title    string
	language string
	content  string
}

func TestImportFromSnippetsLabBuildsTheLibraryAndReportsWhatItLeftBehind(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)

	report, err := app.ImportSnippetsLab.Run(t.Context(), library.ImportSnippetsLabInput{Path: snippetsLabFixture})

	require.NoError(t, err)
	assert.Equal(t, library.Counts{Folders: 3, Tags: 2, Snippets: 2, Fragments: 3}, report.Counts)
	assert.Equal(t, 1, report.Losses.TrashedSnippets)
	assert.Equal(t, []string{"flags"}, titlesOf(report.Losses.LostPins))
	assert.Equal(t, []string{"yacc"}, titlesOf(report.Losses.LostGistLinks))
	assert.Equal(t, []string{"flags"}, titlesOf(report.Losses.LostAttachments))
	require.Len(t, report.Losses.FolderFallbacks, 1)
	assert.Equal(t, "grammars", report.Losses.FolderFallbacks[0].Folder.String())
	require.Len(t, report.Losses.FragmentFallbacks, 1)
	assert.Equal(t, "yacc", report.Losses.FragmentFallbacks[0].Snippet.String())
	assert.Equal(t, 1, report.Losses.FragmentFallbacks[0].Position)

	folders := testapp.StoredFolders(t, app)
	goFolder, cli, grammars := folderNamed(
		t,
		folders,
		"go",
	), folderNamed(
		t,
		folders,
		"cli",
	), folderNamed(
		t,
		folders,
		"grammars",
	)
	assert.True(t, goFolder.AtRoot())
	assert.Equal(t, "Go", goFolder.DefaultLanguage().String())
	assert.Equal(t, goFolder.ID(), cli.ParentID())
	assert.Equal(t, "Go", cli.DefaultLanguage().String())
	assert.Equal(t, "plaintext", grammars.DefaultLanguage().String())

	snippets := testapp.StoredSnippets(t, app)
	flags := snippetTitled(t, snippets, "flags")
	assert.Equal(t, cli.ID(), flags.FolderID())
	assert.Equal(t, "main.go\nParses flags.\n\nBash\nRun with -h.", flags.Description().String())
	assert.Equal(t, []string{"Go", "shell"}, tagNames(flags.Tags()))
	assert.Equal(t, time.Date(2024, time.March, 16, 8, 0, 0, 0, time.UTC), flags.CreatedAt())
	assert.Equal(t, time.Date(2024, time.April, 1, 10, 0, 0, 0, time.UTC), flags.UpdatedAt())
	assert.Equal(t, []fragmentView{
		{title: "main.go", language: "Go", content: "flag.Parse()\n"},
		{title: "", language: "Bash", content: "go run . -h\n"},
	}, fragmentViews(flags))

	yacc := snippetTitled(t, snippets, "yacc")
	assert.Equal(t, grammars.ID(), yacc.FolderID())
	assert.Equal(t, []fragmentView{{title: "", language: "plaintext", content: "%%\n"}}, fragmentViews(yacc))
}

func TestImportFromSnippetsLabReusesAStoredTagOfTheSameName(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	stored := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})

	report, err := app.ImportSnippetsLab.Run(t.Context(), library.ImportSnippetsLabInput{Path: snippetsLabFixture})

	require.NoError(t, err)
	assert.Equal(t, 1, report.Counts.Tags)

	assert.Len(t, testapp.StoredTags(t, app), 2)
	assert.Contains(t, snippetTitled(t, testapp.StoredSnippets(t, app), "flags").Tags(), stored)
}

func TestImportFromSnippetsLabThatFailsLeavesTheLibraryUntouched(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	dir := t.TempDir()
	before := exported(t, app, filepath.Join(dir, "before.json"))

	_, err := app.ImportSnippetsLab.Run(t.Context(), library.ImportSnippetsLabInput{Path: brokenSnippetsLabExport(t)})

	require.ErrorIs(t, err, value.ErrTitleTooLong)
	assert.Equal(t, string(before), string(exported(t, app, filepath.Join(dir, "after.json"))))
}

func brokenSnippetsLabExport(t *testing.T) string {
	t.Helper()

	fixture, err := os.ReadFile(snippetsLabFixture)
	require.NoError(t, err)

	broken := strings.Replace(string(fixture), `"title": "yacc"`, `"title": "`+strings.Repeat("y", 201)+`"`, 1)
	path := filepath.Join(t.TempDir(), "snippetslab.json")
	require.NoError(t, os.WriteFile(path, []byte(broken), 0o600)) //nolint:gosec // G703: a path under t.TempDir

	return path
}

func folderNamed(t *testing.T, folders []domain.Folder, name string) domain.Folder {
	t.Helper()

	return only(t, folders, func(folder domain.Folder) bool { return folder.Name().String() == name })
}

func snippetTitled(t *testing.T, snippets []domain.Snippet, title string) domain.Snippet {
	t.Helper()

	return only(t, snippets, func(snippet domain.Snippet) bool { return snippet.Title().String() == title })
}

func only[T any](t *testing.T, items []T, matches func(T) bool) T {
	t.Helper()

	var found []T

	for _, item := range items {
		if matches(item) {
			found = append(found, item)
		}
	}

	require.Len(t, found, 1)

	return found[0]
}

func fragmentViews(snippet domain.Snippet) []fragmentView {
	views := make([]fragmentView, 0, len(snippet.Fragments()))
	for _, fragment := range snippet.Fragments() {
		views = append(views, fragmentView{
			title:    fragment.Title().String(),
			language: fragment.Language().String(),
			content:  fragment.Content().String(),
		})
	}

	return views
}

func tagNames(tags []domain.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, carried := range tags {
		names = append(names, carried.Name().String())
	}

	return names
}

func titlesOf(titles []value.Title) []string {
	names := make([]string, 0, len(titles))
	for _, title := range titles {
		names = append(names, title.String())
	}

	return names
}
