package editor_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/editor"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	originalContent = "package main\n"
	goFileName      = "snippet.go"
)

func TestLauncher_Open_Resolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		installed  []string
		configured string
		env        []string
		want       string
	}{
		{
			name:       "prefers the configured editor",
			installed:  []string{"configured", "visual", "editor", "nvim"},
			configured: "configured",
			env:        []string{"VISUAL=visual", "EDITOR=editor"},
			want:       "configured",
		},
		{
			name:      "falls back to VISUAL",
			installed: []string{"visual", "editor", "nvim"},
			env:       []string{"VISUAL=visual", "EDITOR=editor"},
			want:      "visual",
		},
		{
			name:       "skips a blank configured editor",
			installed:  []string{"visual"},
			configured: "  ",
			env:        []string{"VISUAL=visual"},
			want:       "visual",
		},
		{
			name:      "falls back to EDITOR",
			installed: []string{"editor", "nvim"},
			env:       []string{"VISUAL=", "EDITOR=editor"},
			want:      "editor",
		},
		{name: "falls back to nvim on PATH", installed: []string{"nvim", "vim", "vi", "nano"}, want: "nvim"},
		{name: "falls back to vim on PATH", installed: []string{"vim", "vi", "nano"}, want: "vim"},
		{name: "falls back to vi on PATH", installed: []string{"vi", "nano"}, want: "vi"},
		{name: "falls back to nano on PATH", installed: []string{"nano"}, want: "nano"},
		{
			name:       "passes the arguments a command carries",
			installed:  []string{"code"},
			configured: "code --wait",
			want:       "code --wait",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			editors := newFakeEditors(t)
			for _, name := range tt.installed {
				editors.install(t, name, signingScript)
			}

			edited, err := edit(t, editors.launcher(tt.configured, tt.env...), goFileName)

			require.NoError(t, err)
			assert.Equal(t, tt.want, edited)
		})
	}
}

func TestLauncher_Open_NoEditor(t *testing.T) {
	t.Parallel()

	editors := newFakeEditors(t)

	_, err := editors.launcher("").Open(t.Context(), goFileName, originalContent)

	require.ErrorIs(t, err, domain.ErrNoEditor)
}

func TestLauncher_Open_Run(t *testing.T) {
	t.Parallel()

	t.Run("reads back what the editor wrote to the content", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)

		edited, err := edit(t, editors.launcher(editors.install(t, "appending", appendingScript)), goFileName)

		require.NoError(t, err)
		assert.Equal(t, originalContent+" edited", edited)
	})

	t.Run("names the file it opens", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)

		_, err := edit(t, editors.launcher(editors.install(t, "signing", signingScript)), "Dockerfile")

		require.NoError(t, err)
		assert.Equal(t, "Dockerfile", filepath.Base(editors.editedPath(t)))
	})

	t.Run("keeps the file in a private directory under TMPDIR", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)
		launcher := editors.launcher(editors.install(t, "checking", permissionCheckerScript))

		edited, err := edit(t, launcher, goFileName)

		require.NoError(t, err)
		assert.Equal(t, "private", edited)
	})

	t.Run("removes the directory after reading the file back", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)

		_, err := edit(t, editors.launcher(editors.install(t, "signing", signingScript)), goFileName)

		require.NoError(t, err)
		assertDirRemoved(t, editors.editedPath(t))
	})

	t.Run("discards the content and removes the directory when the editor fails", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)

		edited, err := edit(t, editors.launcher(editors.install(t, "failing", failingScript)), goFileName)

		require.ErrorIs(t, err, domain.ErrEditorFailed)
		assert.Empty(t, edited)
		assertDirRemoved(t, editors.editedPath(t))
	})

	t.Run("reports an editor that could not start", func(t *testing.T) {
		t.Parallel()

		editors := newFakeEditors(t)
		launcher := editors.launcher(editors.install(t, "signing", signingScript))
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		run, err := launcher.Open(ctx, goFileName, originalContent)
		require.NoError(t, err)

		_, err = run(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})

		require.ErrorAs(t, err, &domain.EditorStartError{})
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func edit(t *testing.T, launcher *editor.Launcher, fileName string) (string, error) {
	t.Helper()

	run, err := launcher.Open(t.Context(), fileName, originalContent)
	require.NoError(t, err)

	return run(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
}

func assertDirRemoved(t *testing.T, editedPath string) {
	t.Helper()

	assert.True(t, strings.HasPrefix(editedPath, os.TempDir()), "%s is not under %s", editedPath, os.TempDir())
	assert.NoDirExists(t, filepath.Dir(editedPath))
}
