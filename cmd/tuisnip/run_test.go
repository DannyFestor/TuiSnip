package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type outcome struct {
	code   int
	stdout string
	stderr string
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("prints the paths and creates nothing", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()

		got := runWith(t, home, "--paths")

		assert.Equal(t, exitOK, got.code)
		assert.Equal(t, ""+
			"config   "+home+"/.config/tuisnip/config.toml\n"+
			"data     "+home+"/.local/share/tuisnip/tuisnip.db\n"+
			"backup   "+home+"/.local/share/tuisnip/backups\n"+
			"state    "+home+"/.local/state/tuisnip/state.toml\n"+
			"log      "+home+"/.local/state/tuisnip/tuisnip.log\n",
			got.stdout)
		assert.Empty(t, entries(t, home))
	})

	t.Run("prints the version", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "--version")

		assert.Equal(t, exitOK, got.code)
		assert.Regexp(t, `^tuisnip \S+\n$`, got.stdout)
	})

	t.Run("prints usage with the flags and commands for help", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "-h")

		assert.Equal(t, exitOK, got.code)
		assert.Contains(t, got.stderr, "-paths")
		assert.Contains(t, got.stderr, "export <file>")
		assert.Contains(t, got.stderr, "import <file>")
		assert.Contains(t, got.stderr, "import --from snippetslab <file>")
	})

	t.Run("refuses an unknown flag", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "--sync")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "flag provided but not defined: -sync")
	})

	t.Run("refuses an unknown command and prints usage", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()

		got := runWith(t, home, "get", "curl")

		assert.Equal(t, exitUsage, got.code)
		assert.Contains(t, got.stderr, "tuisnip: unknown command: get\n")
		assert.Contains(t, got.stderr, "export <file>")
		assert.Empty(t, entries(t, home))
	})

	t.Run("exports the Library and reports the path and counts", func(t *testing.T) {
		t.Parallel()

		target := filepath.Join(t.TempDir(), "backup.json")

		got := runWith(t, t.TempDir(), "export", target)

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t, "Exported 0 Folders, 0 Tags, 0 Snippets to "+target+"\n", got.stdout)
		assert.Contains(t, read(t, target), `"format": "tuisnip-export"`)
	})

	t.Run("imports an Export into an empty Library and reports the counts", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "import", exportHoldingOneTag(t))

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t, "Imported 0 Folders, 1 Tag, 0 Snippets\n", got.stdout)
	})

	t.Run("imports an Export holding the Trash and reports the Trash on a second line", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "import", exportHoldingOnlyTrash(t))

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t, "Imported 1 Folder, 0 Tags, 2 Snippets\nIn the Trash: 1 Folder, 2 Snippets\n", got.stdout)
	})

	t.Run("exports the Trash and reports it on a second line", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		require.Equal(t, exitOK, runWith(t, home, "import", exportHoldingOnlyTrash(t)).code)
		target := filepath.Join(t.TempDir(), "backup.json")

		got := runWith(t, home, "export", target)

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t,
			"Exported 1 Folder, 0 Tags, 2 Snippets to "+target+"\nIn the Trash: 1 Folder, 2 Snippets\n",
			got.stdout)
	})

	t.Run("refuses to import into a Library holding only Trash items", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		require.Equal(t, exitOK, runWith(t, home, "import", exportHoldingOnlyTrash(t)).code)

		got := runWith(t, home, "import", exportHoldingOneTag(t))

		assert.Equal(t, exitFailure, got.code)
		assert.Empty(t, got.stdout)
		assert.Contains(t, got.stderr, "tuisnip --paths")
	})

	t.Run("imports a SnippetsLab export and prints the closing report", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, t.TempDir(), "import", "--from", "snippetslab", snippetsLabExport(t))

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.Equal(t, ""+
			"Imported 1 Folder, 1 Tag, 2 Snippets, 4 Fragments\n"+
			"Skipped 1 Snippet in the Trash\n"+
			"Lost a pin: loop; parser\n"+
			"Lost a Gist link: loop\n"+
			"Lost an attachment: parser\n"+
			"Fell back to plaintext:\n"+
			"  Folder grammars: YaccLexer\n"+
			"  Snippet parser, Fragment 2 (grammar.y): YaccLexer\n"+
			"  Snippet parser, Fragment 3: no lexer\n",
			got.stdout)
	})

	t.Run("imports a second SnippetsLab export beside the first", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		path := snippetsLabExport(t)
		require.Equal(t, exitOK, runWith(t, home, "import", "--from", "snippetslab", path).code)

		got := runWith(t, home, "import", "--from", "snippetslab", path)

		assert.Equal(t, exitOK, got.code, got.stderr)
		assert.True(
			t,
			strings.HasPrefix(got.stdout, "Imported 1 Folder, 0 Tags, 2 Snippets, 4 Fragments\n"),
			got.stdout,
		)
	})

	t.Run("refuses to import into a Library that isn't empty and points to --paths", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		path := exportHoldingOneTag(t)
		require.Equal(t, exitOK, runWith(t, home, "import", path).code)

		got := runWith(t, home, "import", path)

		assert.Equal(t, exitFailure, got.code)
		assert.Empty(t, got.stdout)
		assert.True(t, strings.HasPrefix(got.stderr, "tuisnip: import: "), got.stderr)
		assert.Contains(t, got.stderr, "tuisnip --paths")
	})

	refusals := []struct {
		name       string
		args       []string
		wantStderr []string
	}{
		{
			name:       "refuses an export without a file",
			args:       []string{"export"},
			wantStderr: []string{"tuisnip: export: wrong number of arguments\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses an export to two files",
			args:       []string{"export", "one.json", "two.json"},
			wantStderr: []string{"tuisnip: export: wrong number of arguments\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses - as the file to export to",
			args:       []string{"export", "-"},
			wantStderr: []string{"tuisnip: export: - is not supported\n", "usage: tuisnip export <file>\n"},
		},
		{
			name:       "refuses an import without a file",
			args:       []string{"import"},
			wantStderr: []string{"tuisnip: import: wrong number of arguments\n", "usage: tuisnip import <file>\n"},
		},
		{
			name: "refuses an import from an app it can't read",
			args: []string{"import", "--from", "bear", "notes.json"},
			wantStderr: []string{
				"tuisnip: import: unknown source: bear\n",
				"usage: tuisnip import <file>\n",
				"usage: tuisnip import --from snippetslab <file>\n",
			},
		},
		{
			name:       "refuses --from on a command that reads no other app's file",
			args:       []string{"export", "--from", "snippetslab", "backup.json"},
			wantStderr: []string{"flag provided but not defined: -from\n"},
		},
		{
			name:       "refuses --paths together with a command",
			args:       []string{"--paths", "export", "backup.json"},
			wantStderr: []string{"tuisnip: -version and -paths take no command\n", "export <file>"},
		},
		{
			name:       "refuses --version together with a command",
			args:       []string{"--version", "export", "backup.json"},
			wantStderr: []string{"tuisnip: -version and -paths take no command\n", "export <file>"},
		},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			home := t.TempDir()

			got := runWith(t, home, tt.args...)

			assert.Equal(t, exitUsage, got.code)
			assert.Empty(t, got.stdout)
			assert.Empty(t, entries(t, home))

			for _, want := range tt.wantStderr {
				assert.Contains(t, got.stderr, want)
			}
		})
	}

	t.Run("leaves an existing file untouched when the export fails", func(t *testing.T) {
		t.Parallel()

		if os.Geteuid() == 0 {
			t.Skip("root writes into a read-only directory")
		}

		dir := t.TempDir()
		target := filepath.Join(dir, "backup.json")
		require.NoError(t, os.WriteFile(target, []byte("good backup"), 0o600))
		require.NoError(t, os.Chmod(dir, 0o500)) //nolint:gosec // G302: a directory needs its execute bit
		t.Cleanup(func() {
			require.NoError(t, os.Chmod(dir, 0o700)) //nolint:gosec // G302: a directory needs its execute bit
		})

		got := runWith(t, t.TempDir(), "export", target)

		assert.Equal(t, exitFailure, got.code)
		assert.True(t, strings.HasPrefix(got.stderr, "tuisnip: export: "), got.stderr)
		assert.Empty(t, got.stdout)
		assert.Equal(t, "good backup", read(t, target))
		assert.Len(t, entries(t, dir), 1)
	})

	t.Run("fails without a HOME", func(t *testing.T) {
		t.Parallel()

		got := runWith(t, "", "--paths")

		assert.Equal(t, exitFailure, got.code)
		assert.Contains(t, got.stderr, "tuisnip: xdg: HOME is not set")
	})

	t.Run("refuses to start with an invalid config and names the key", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		configFile := home + "/.config/tuisnip/config.toml"
		require.NoError(t, os.MkdirAll(filepath.Dir(configFile), 0o700))
		require.NoError(t, os.WriteFile(configFile, []byte("theme = \"purple\"\n"), 0o600))

		got := runWith(t, home)

		assert.Equal(t, exitFailure, got.code)
		assert.True(t, strings.HasPrefix(got.stderr, "tuisnip: invalid config "+configFile+":\n"), got.stderr)
		assert.Contains(t, got.stderr, "theme")
	})
}

func runWith(t *testing.T, home string, args ...string) outcome {
	t.Helper()

	var stdout, stderr bytes.Buffer

	code := run(t.Context(), args, environment{
		stdout:   &stdout,
		stderr:   &stderr,
		getenv:   homeOnly(home),
		environ:  func() []string { return nil },
		lookPath: func(string) (string, error) { return "", exec.ErrNotFound },
		goos:     "darwin",
	})

	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func homeOnly(home string) func(string) string {
	return func(key string) string {
		if key == "HOME" {
			return home
		}

		return ""
	}
}

func exportHoldingOneTag(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "backup.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
  "format": "tuisnip-export",
  "version": 1,
  "folders": [],
  "tags": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000002",
      "name": "Shell",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z"
    }
  ],
  "snippets": []
}
`), 0o600))

	return path
}

func exportHoldingOnlyTrash(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "backup.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
  "format": "tuisnip-export",
  "version": 2,
  "folders": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000001",
      "parent_id": null,
      "name": "scripts",
      "default_language": "Bash",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z",
      "deleted_at": "2026-03-03T09:00:00.000000000Z",
      "former_parent_path": []
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
      ],
      "deleted_at": null,
      "former_folder_path": []
    },
    {
      "id": "0194c3a0-0000-7000-8000-000000000005",
      "folder_id": null,
      "title": "note",
      "description": "",
      "tag_ids": [],
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z",
      "fragments": [
        {
          "id": "0194c3a0-0000-7000-8000-000000000006",
          "title": "",
          "language": "plaintext",
          "content": "remember",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-01T09:30:00.123456789Z"
        }
      ],
      "deleted_at": "2026-03-02T09:00:00.000000000Z",
      "former_folder_path": ["gone"]
    }
  ]
}
`), 0o600))

	return path
}

func snippetsLabExport(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "snippetslab.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
  "contents": {
    "folders": [{"uuid": "F1", "title": "grammars", "defaultLanguage": "YaccLexer", "children": []}],
    "tags": [{"uuid": "T1", "title": "Go"}],
    "snippets": [
      {
        "uuid": "S1", "title": "loop", "pinned": true, "tags": ["T1"],
        "gistIdentifier": "0123456789abcdef0123456789abcdef",
        "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z",
        "fragments": [{"title": "Fragment", "language": "GoLexer", "content": "for {}",
          "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z"}]
      },
      {
        "uuid": "S2", "title": "parser", "pinned": true, "folder": "F1",
        "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z",
        "fragments": [
          {"title": "main.go", "language": "GoLexer", "content": "package main",
            "attachments": [{"uuid": "A1", "file": {"filename": "diagram.png", "bytes": 2048}}],
            "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z"},
          {"title": "grammar.y", "language": "YaccLexer", "content": "%%",
            "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z"},
          {"content": "notes",
            "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z"}
        ]
      },
      {
        "uuid": "S3", "title": "old", "dateDeleted": "2024-06-01T00:00:00Z",
        "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z",
        "fragments": [{"content": "gone",
          "dateCreated": "2024-05-01T09:30:00Z", "dateModified": "2024-05-01T09:30:00Z"}]
      }
    ]
  }
}
`), 0o600))

	return path
}

func read(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: a path under t.TempDir
	require.NoError(t, err)

	return string(data)
}

func entries(t *testing.T, dir string) []os.DirEntry {
	t.Helper()

	found, err := os.ReadDir(dir)
	require.NoError(t, err)

	return found
}
