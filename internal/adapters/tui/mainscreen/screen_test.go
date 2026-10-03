package mainscreen_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestScreen_navigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		keys  []tea.KeyPressMsg
		title string
	}{
		{name: "starts on Folders", keys: nil, title: folderPaneTitle},
		{name: "tab moves to Tags", keys: []tea.KeyPressMsg{keypress.Special(tea.KeyTab)}, title: tagPaneTitle},
		{
			name:  "tab wraps from the Snippet pane to Folders",
			keys:  []tea.KeyPressMsg{keypress.Letter('4'), keypress.Special(tea.KeyTab)},
			title: folderPaneTitle,
		},
		{
			name:  "shift+tab wraps from Folders to the Snippet pane",
			keys:  []tea.KeyPressMsg{{Code: tea.KeyTab, Mod: tea.ModShift}},
			title: snippetPaneTitle,
		},
		{name: "l from Tags goes to the Snippet list", keys: keypress.Typed("2l"), title: snippetListTitle},
		{name: "h from the Snippet list returns to Folders", keys: keypress.Typed("2lh"), title: folderPaneTitle},
		{name: "l stops at the Snippet pane", keys: keypress.Typed("4l"), title: snippetPaneTitle},
		{name: "h stops at the left column", keys: keypress.Typed("2h"), title: tagPaneTitle},
		{
			name:  "enter drills from Folders to the Snippet pane",
			keys:  []tea.KeyPressMsg{keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEnter)},
			title: snippetPaneTitle,
		},
		{
			name:  "enter does nothing in Tags",
			keys:  []tea.KeyPressMsg{keypress.Letter('2'), keypress.Special(tea.KeyEnter)},
			title: tagPaneTitle,
		},
		{
			name: "esc backs out from the Snippet pane to Folders",
			keys: []tea.KeyPressMsg{
				keypress.Letter('4'),
				keypress.Special(tea.KeyEscape),
				keypress.Special(tea.KeyEscape),
			},
			title: folderPaneTitle,
		},
		{name: "3 focuses the Snippet list", keys: keypress.Typed("3"), title: snippetListTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := showing(t, narrow())

			screen.Press(tt.keys...)

			assert.Contains(t, screen.Screen(), tt.title)
		})
	}
}

func TestScreen_cursor(t *testing.T) {
	t.Parallel()

	t.Run("moving down in the Snippet list shows the next Snippet", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("3j")...)

		assert.Contains(t, screen.Screen(), "Reclaim disk space")
		assert.NotContains(t, screen.Screen(), firstDescription)
	})

	t.Run("moving down in Folders keeps the Snippet", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('j'))

		assert.Contains(t, screen.Screen(), firstDescription)
	})

	t.Run("selects the Snippet it was loaded with", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide())

		screen.Send(mainscreen.SnippetsLoaded{Snippets: snippets, Selecting: snippets[1].ID()})

		assert.Contains(t, screen.Screen(), "Reclaim disk space")
	})
}

func TestScreen_layout(t *testing.T) {
	t.Parallel()

	t.Run("shows all four Panes at the minimum size", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum())

		for _, title := range []string{folderPaneTitle, tagPaneTitle, snippetListTitle, snippetPaneTitle} {
			assert.Contains(t, screen.Screen(), title)
		}
	})

	t.Run("shows only the focused Pane below the minimum size", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum())

		screen.Send(tea.WindowSizeMsg{Width: minimum().Width - 1, Height: minimum().Height})

		assert.Contains(t, screen.Screen(), folderPaneTitle)
		assert.NotContains(t, screen.Screen(), tagPaneTitle)
	})

	widths := []struct {
		name    string
		focus   rune
		columns []int
	}{
		{name: "widens the left column when Folders has focus", focus: '1', columns: []int{34, 36, 50}},
		{name: "widens the left column when Tags has focus", focus: '2', columns: []int{34, 36, 50}},
		{name: "widens the Snippet list when it has focus", focus: '3', columns: []int{26, 44, 50}},
		{name: "widens the Snippet pane when it has focus", focus: '4', columns: []int{22, 32, 66}},
	}

	for _, tt := range widths {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := showing(t, wide())

			screen.Press(keypress.Letter(tt.focus))

			assert.Equal(t, tt.columns, columnWidths(screen))
		})
	}

	heights := []struct {
		name   string
		focus  rune
		tagRow int
	}{
		{name: "makes Folders the tall left Pane when it has focus", focus: '1', tagRow: 26},
		{name: "makes Tags the tall left Pane when it has focus", focus: '2', tagRow: 13},
		{name: "keeps the selection holder tall outside the left column", focus: '3', tagRow: 26},
	}

	for _, tt := range heights {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := showing(t, wide())

			screen.Press(keypress.Letter('2'), keypress.Letter(tt.focus))

			assert.Equal(t, tt.tagRow, lineIndexOf(screen, tagPaneTitle))
		})
	}

	t.Run("draws the focused Pane in the focused style", func(t *testing.T) {
		t.Parallel()

		screen := showingStyled(t, wide(), upperFocusedTitle())

		screen.Press(keypress.Letter('3'))

		assert.Contains(t, screen.Screen(), strings.ToUpper(snippetListTitle))
		assert.Contains(t, screen.Screen(), folderPaneTitle)
	})
}

func TestScreen_statusLine(t *testing.T) {
	t.Parallel()

	t.Run("shows the focused Pane's hints on the right", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('3'))

		assert.True(t, strings.HasSuffix(statusLine(screen), " "+listHint))
	})

	t.Run("shows the status text on the left", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Send(mainscreen.StatusShown{Text: "Copied"})

		assert.True(t, strings.HasPrefix(statusLine(screen), " Copied "))
	})

	t.Run("cuts a status text too long to fit beside the hints", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('3'))

		room := wide().Width - ansi.StringWidth(listHint) - 1
		screen.Send(mainscreen.StatusShown{Text: strings.Repeat("x", room)})

		assert.Equal(t, " "+strings.Repeat("x", room-2)+"… "+listHint, statusLine(screen))
	})

	t.Run("keeps the hints that fit in half the width", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('/'))

		assert.True(t, strings.HasSuffix(statusLine(screen), " down move · enter reveal · ctrl+y Copy"))
	})

	t.Run("keeps hints exactly half the width wide", func(t *testing.T) {
		t.Parallel()

		hints := "down move · enter reveal · ctrl+y Copy"
		screen := showing(t, minimum(), sampleSnippets(t)...)
		screen.Send(tea.WindowSizeMsg{Width: 2 * ansi.StringWidth(hints), Height: minimum().Height})

		screen.Press(keypress.Letter('/'))

		assert.True(t, strings.HasSuffix(statusLine(screen), " "+hints))
	})

	t.Run("says the terminal is too small below the minimum size", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, narrow())

		assert.Equal(
			t,
			strings.Repeat(" ", narrow().Width-ansi.StringWidth(tooSmallHint))+tooSmallHint,
			statusLine(screen),
		)
	})

	t.Run("shows an open Overlay's hints instead of the too-small hint", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, narrow(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('/'))

		assert.NotContains(t, statusLine(screen), tooSmallHint)
		assert.Contains(t, statusLine(screen), "down move")
	})
}

func TestScreen_Update(t *testing.T) {
	t.Parallel()

	t.Run("q asks to quit", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('q'))

		assert.Equal(t, []outcome.Outcome{outcome.QuitAsked{}}, screen.Outcomes())
	})

	t.Run("n opens the edit overlay", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('n'))

		assert.Contains(t, screen.Screen(), "Editing")
		assert.Equal(t, "ctrl+s save · esc cancel · down field", screen.Hints())
	})

	t.Run("/ opens the Search popup over the listed Snippets", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('/'))

		assert.Contains(t, screen.Screen(), "Search · 2 results")
	})

	t.Run("y in the Snippet list asks to copy the selected Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("3y")...)

		assert.Equal(t, []outcome.Outcome{outcome.CopyRequested{ID: snippets[0].ID()}}, screen.Outcomes())
	})

	t.Run("y in the Snippet pane asks to copy the shown Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("4y")...)

		assert.Equal(t, []outcome.Outcome{outcome.CopyRequested{ID: snippets[0].ID()}}, screen.Outcomes())
	})

	t.Run("y does nothing in Folders and Tags", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("y2y")...)

		assert.Empty(t, screen.Outcomes())
	})
}

func TestScreen_Received(t *testing.T) {
	t.Parallel()

	t.Run("focuses the Snippet pane on a revealed Snippet and passes it on", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		revealed := outcome.SnippetRevealed{ID: snippets[1].ID()}
		screen := showing(t, wide(), snippets...)

		screen.Offer(revealed)

		assert.Equal(t, []outcome.Outcome{revealed}, screen.Outcomes())
		assert.Equal(t, []int{22, 32, 66}, columnWidths(screen))
	})

	t.Run("passes any other outcome on", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Offer(outcome.QuitAsked{})

		assert.Equal(t, []outcome.Outcome{outcome.QuitAsked{}}, screen.Outcomes())
	})
}

func TestScreen_FullHelp(t *testing.T) {
	t.Parallel()

	keys := testsettings.Default(t).Keys
	screen := mainscreen.New(keys, look.NewStyles(), time.UTC)

	assert.Equal(t, folderpane.New(keys).FullHelp(), screen.FullHelp())
}
