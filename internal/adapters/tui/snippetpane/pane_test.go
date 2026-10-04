package snippetpane_test

import (
	"image/color"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the empty hint with no Snippet", func(t *testing.T) {
		t.Parallel()

		got := lines(paneIn(t, time.UTC, look.Size{Width: boxWidth, Height: headerLines}))

		assert.Equal(t, "No Snippets here.", got[0])
		assert.Contains(t, got, "n  new Snippet")
	})

	t.Run("shows the header above the numbered Fragment", func(t *testing.T) {
		t.Parallel()

		got := lines(showing(t, longSnippet(t)))

		assert.Equal(t, []string{
			"Long",
			"Root · Go",
			"created 2026-09-06 · updated 2026-09-06",
			"────────────────────────",
			"   1 │ line 1",
			"   2 │ line 2",
			"   3 │ line 3",
		}, trimmed(got))
	})

	t.Run("shows the Description under the dates and gives the Fragment less room", func(t *testing.T) {
		t.Parallel()

		pane := showing(t, snippetWith(t, testkit.SnippetSpec{Description: "first\nsecond"}))

		assert.Equal(t, []string{
			"Long",
			"Root · Go",
			"created 2026-09-06 · updated 2026-09-06",
			"first",
			"second",
			"────────────────────────",
			"   1 │ line 1",
		}, trimmed(lines(pane)))
	})

	t.Run("shows the Folder path of a Snippet filed in a Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		pane := paneIn(t, time.UTC, look.Size{Width: boxWidth, Height: headerLines + codeLines}).
			WithPaths(folderpath.New(sample.Tree)).
			Showing(snippetWith(t, testkit.SnippetSpec{FolderID: sample.Tests.ID()}))

		assert.Equal(t, "Root / go / testing · Go", strings.TrimSpace(lines(pane)[1]))
	})

	t.Run("shows the dates in the configured location", func(t *testing.T) {
		t.Parallel()

		ahead := time.FixedZone("ahead", int((15 * time.Hour).Seconds()))
		pane := paneIn(t, ahead, look.Size{Width: boxWidth, Height: headerLines + codeLines}).Showing(longSnippet(t))

		assert.Equal(t, "created 2026-09-07 · updated 2026-09-07", lines(pane)[2])
	})
}

func TestPane_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []tea.KeyPressMsg
		want string
	}{
		{name: "down scrolls one line", keys: []tea.KeyPressMsg{keypress.Letter('j')}, want: "   2 │ line 2"},
		{
			name: "up scrolls back one line",
			keys: []tea.KeyPressMsg{keypress.Letter('j'), keypress.Letter('k')},
			want: "   1 │ line 1",
		},
		{name: "G scrolls to the end", keys: []tea.KeyPressMsg{keypress.Letter('G')}, want: "   8 │ line 8"},
		{
			name: "g scrolls to the top",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Letter('g')},
			want: "   1 │ line 1",
		},
		{
			name: "page down scrolls one page",
			keys: []tea.KeyPressMsg{keypress.Special(tea.KeyPgDown)},
			want: "   4 │ line 4",
		},
		{
			name: "page up scrolls back one page",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Special(tea.KeyPgUp)},
			want: "   5 │ line 5",
		},
		{name: "ignores keys without a Binding", keys: []tea.KeyPressMsg{keypress.Letter('x')}, want: "   1 │ line 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pane := pressed(t, showing(t, longSnippet(t)), tt.keys...)

			assert.Equal(t, tt.want, trimmed(code(pane))[0])
		})
	}

	t.Run("asks to Copy the shown Snippet", func(t *testing.T) {
		t.Parallel()

		snippet := longSnippet(t)

		_, outcomes, _ := showing(t, snippet).Update(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.CopyRequested{ID: snippet.ID()}}, outcomes)
	})

	t.Run("asks for no Copy with no Snippet", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := showing(t, longSnippet(t)).Cleared().Update(keypress.Letter('y'))

		assert.Empty(t, outcomes)
	})

	t.Run("asks for no Copy on other keys", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := showing(t, longSnippet(t)).Update(keypress.Letter('j'))

		assert.Empty(t, outcomes)
	})

	t.Run("highlights for the terminal's background", func(t *testing.T) {
		t.Parallel()

		pane := showing(t, longSnippet(t))

		light, _, _ := pane.Update(tea.BackgroundColorMsg{Color: color.White})

		assert.NotEqual(t, pane.View(), light.View())
		assert.Equal(t, lines(pane), lines(light))
	})

	t.Run("ignores messages it does not handle", func(t *testing.T) {
		t.Parallel()

		pane := showing(t, longSnippet(t))

		next, outcomes, cmd := pane.Update(tea.FocusMsg{})

		assert.Equal(t, pane.View(), next.View())
		assert.Empty(t, outcomes)
		assert.Nil(t, cmd)
	})
}

func TestPane_Showing(t *testing.T) {
	t.Parallel()

	t.Run("keeps the scroll while showing the same Snippet", func(t *testing.T) {
		t.Parallel()

		snippet := longSnippet(t)
		pane := pressed(t, showing(t, snippet), keypress.Letter('j'))

		assert.Equal(t, "   2 │ line 2", trimmed(code(pane.Showing(snippet)))[0])
	})

	t.Run("scrolls back to the top for a changed Snippet", func(t *testing.T) {
		t.Parallel()

		pane := pressed(t, showing(t, longSnippet(t)), keypress.Letter('j'))
		changed := snippetWith(t, testkit.SnippetSpec{UpdatedAt: created().AddDate(0, 0, 1)})

		assert.Equal(t, "   1 │ line 1", trimmed(code(pane.Showing(changed)))[0])
	})

	t.Run("shows a Snippet again after being cleared", func(t *testing.T) {
		t.Parallel()

		snippet := longSnippet(t)

		pane := showing(t, snippet).Cleared().Showing(snippet)

		assert.Equal(t, "Long", lines(pane)[0])
	})
}

func TestPane_ShortHelp(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, time.UTC, look.Size{Width: boxWidth, Height: headerLines})

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeSnippetPane).ShortHelp(), pane.ShortHelp())
	assert.Equal(t, [][]key.Binding{pane.ShortHelp()}, pane.FullHelp())
}
