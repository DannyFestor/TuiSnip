package folderoutline_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderoutline"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the Folder's name, place, counts, deletion, and its subtree outline", func(t *testing.T) {
		t.Parallel()

		got := lines(paneShowing(t, 20))

		assert.Equal(t, []string{
			"docker",
			"Root / ops",
			"1 subfolder · 3 Snippets",
			"deleted 2026-10-14",
			"────────────────────────────────────────",
			"Build",
			"zap images",
			"▾ compose",
			"  Up",
		}, got)
	})
}

func TestPane_Update(t *testing.T) {
	t.Parallel()

	t.Run("scrolls the outline with the movement keys", func(t *testing.T) {
		t.Parallel()

		pane, _, _ := paneShowing(t, 7).Update(keypress.Letter('G'))

		assert.Equal(t, []string{"▾ compose", "  Up"}, lines(pane)[5:])
	})
}

func paneShowing(t *testing.T, height int) folderoutline.Pane {
	t.Helper()

	pane := folderoutline.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark))
	pane, _, _ = pane.Update(look.Resized{Box: look.Size{Width: 40, Height: height}})

	docker := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "docker"})

	return pane.Showing(trash.Item{
		Folder: docker,
		Contents: trash.Contents{
			SubfolderCount: 1,
			SnippetCount:   3,
			Outline: trash.Outline{
				Name:          "docker",
				SnippetTitles: []string{"Build", "zap images"},
				Subfolders:    []trash.Outline{{Name: "compose", SnippetTitles: []string{"Up"}}},
			},
		},
	}, "Root / ops", "deleted 2026-10-14")
}

func lines(pane folderoutline.Pane) []string {
	split := strings.Split(ansi.Strip(pane.View()), "\n")
	trimmed := make([]string, 0, len(split))

	for _, line := range split {
		if text := strings.TrimRight(line, " "); text != "" {
			trimmed = append(trimmed, text)
		}
	}

	return trimmed
}
