package searchpopup_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestPopup_View(t *testing.T) {
	t.Parallel()

	t.Run("lists the Browse selection and previews the first", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		assert.Contains(t, screen.Screen(), "Search · 2 results")
		assert.Contains(t, screen.Screen(), "Preview")
		assert.Contains(t, screen.Screen(), firstContent)
	})

	t.Run("says when nothing matches", func(t *testing.T) {
		t.Parallel()

		screen := searchingIn(t, nil)

		assert.Contains(t, screen.Screen(), "Search · 0 results")
		assert.Contains(t, screen.Screen(), "No Snippets match.")
		assert.Contains(t, screen.Screen(), "No Snippets here.")
	})

	t.Run("shows each result's Folder path without the Root", func(t *testing.T) {
		t.Parallel()

		screen, _ := searchingInFolder(t)

		assert.Regexp(t, `Table test skeleton +go / testing`, screen.Screen())
		assert.Contains(t, screen.Screen(), "Root / go / testing · plaintext")
	})

	t.Run("shows the Folder paths once the Folder tree loads", func(t *testing.T) {
		t.Parallel()

		screen, sample := searchingBeforeTreeLoads(t)
		require.Contains(t, screen.Screen(), "Root / … · plaintext")

		screen.Send(arrived.Tree{Tree: sample.Tree})

		assert.Regexp(t, `Table test skeleton +go / testing`, screen.Screen())
		assert.Contains(t, screen.Screen(), "Root / go / testing · plaintext")
	})

	t.Run("shows Root for a result at the Root", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		assert.Regexp(t, `Graceful HTTP shutdown +Root`, screen.Screen())
	})

	t.Run("splits its share of the screen between results and preview", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		assert.Equal(t, popupWidth, screen.TopBorderWidth())
	})

	t.Run("scrolls a long query to keep its end and the cursor in the frame", func(t *testing.T) {
		t.Parallel()

		query := "abcdefghijklmnopqrstuvwxyz0123456789"
		screen := searching(t)

		screen.Press(keypress.Typed(query)...)

		assert.Contains(t, screen.Screen(), "│/ "+query[len(query)-queryInputWidth:]+" │")
	})

	t.Run("scrolls the results below the query", func(t *testing.T) {
		t.Parallel()

		screen := searchingIn(t, numberedSnippets(t, resultRows+1))

		for range resultRows {
			screen.Press(keypress.Special(tea.KeyDown))
		}

		assert.Contains(t, screen.Screen(), numberedTitle(resultRows+1))
		assert.NotContains(t, screen.Screen(), numberedTitle(1))
	})

	t.Run("draws the results and the preview in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)
		screen := searching(t)

		screen.Send(look.Restyled{Styles: light})

		listing := searchpopup.Listing{Snippets: sampleSnippets(t), Paths: folderpath.Paths{}}
		assert.Equal(t, searchingStyled(t, light, listing).StyledScreen(), screen.StyledScreen())
	})
}

func TestPopup_ShortHelp(t *testing.T) {
	t.Parallel()

	screen := searching(t)

	assert.Equal(t, "down move · enter reveal · ctrl+y Copy · esc close", screen.Hints())
}

func TestPopup_search(t *testing.T) {
	t.Parallel()

	t.Run("asks to search for the query", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Press(keypress.Letter('p'))

		assert.Equal(t, []outcome.Outcome{outcome.SearchTyped{Text: "p"}}, screen.Outcomes())
	})

	t.Run("lists the hits for the query and previews the first", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Letter('p'))
		screen.Send(searchpopup.HitsFound{Text: "p", Hits: hitsOf(snippets[1])})

		assert.Contains(t, screen.Screen(), "Search · 1 result")
		assert.Contains(t, screen.Screen(), secondContent)
	})

	t.Run("ignores hits for an earlier query", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Letter('p'), keypress.Letter('r'))
		screen.Send(searchpopup.HitsFound{Text: "p", Hits: hitsOf(snippets[1])})

		assert.Contains(t, screen.Screen(), "Search · 2 results")
	})

	t.Run("searches a pasted query", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Send(tea.PasteMsg{Content: "prune"})

		assert.Equal(t, []outcome.Outcome{outcome.SearchTyped{Text: "prune"}}, screen.Outcomes())
	})

	t.Run("types Pane keys into the query instead of acting on them", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Press(keypress.Letter('q'))

		assert.Equal(t, []outcome.Outcome{outcome.SearchTyped{Text: "q"}}, screen.Outcomes())
		assert.True(t, screen.IsOpen())
	})

	t.Run("lists the Browse selection again for a blank query without searching", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Letter('p'))
		screen.Send(searchpopup.HitsFound{Text: "p", Hits: hitsOf(snippets[1])})
		screen.Press(keypress.Special(tea.KeyBackspace), keypress.Letter(' '))

		assert.Equal(t, []outcome.Outcome{outcome.SearchTyped{Text: "p"}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "Search · 2 results")
		assert.Contains(t, screen.Screen(), firstContent)
	})

	t.Run("asks nothing for a key that leaves the query as it was", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Press(keypress.Special(tea.KeyLeft))

		assert.Empty(t, screen.Outcomes())
	})
}

func TestPopup_move(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []tea.KeyPressMsg
		want string
	}{
		{
			name: "down previews the next result",
			keys: []tea.KeyPressMsg{keypress.Special(tea.KeyDown)},
			want: secondContent,
		},
		{name: "ctrl+n previews the next result", keys: []tea.KeyPressMsg{keypress.Ctrl('n')}, want: secondContent},
		{
			name: "up stops at the first result",
			keys: []tea.KeyPressMsg{
				keypress.Special(tea.KeyDown),
				keypress.Special(tea.KeyUp),
				keypress.Special(tea.KeyUp),
			},
			want: firstContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := searching(t)

			screen.Press(tt.keys...)

			assert.Contains(t, screen.Screen(), tt.want)
		})
	}

	t.Run("a new list starts at its first result", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Special(tea.KeyDown), keypress.Letter('e'))
		screen.Send(searchpopup.HitsFound{Text: "e", Hits: hitsOf(snippets...)})

		assert.Contains(t, screen.Screen(), firstContent)
	})
}

func TestPopup_end(t *testing.T) {
	t.Parallel()

	t.Run("enter reveals the selected Snippet and closes", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.SnippetRevealed{ID: snippets[1].ID()}}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("enter reveals a Snippet in its Folder", func(t *testing.T) {
		t.Parallel()

		screen, filed := searchingInFolder(t)

		screen.Press(keypress.Special(tea.KeyEnter))

		revealed := outcome.SnippetRevealed{ID: filed.ID(), FolderID: filed.FolderID()}
		assert.Equal(t, []outcome.Outcome{revealed}, screen.Outcomes())
	})

	t.Run("ctrl+y copies the selected Snippet and closes", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := searching(t)

		screen.Press(keypress.Special(tea.KeyDown), keypress.Ctrl('y'))

		assert.Equal(t, []outcome.Outcome{outcome.CopyRequested{ID: snippets[1].ID()}}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("esc closes without an outcome", func(t *testing.T) {
		t.Parallel()

		screen := searching(t)

		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyEscape))

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	for name, pressed := range map[string]tea.KeyPressMsg{"enter": keypress.Special(tea.KeyEnter), "ctrl+y": keypress.Ctrl('y')} {
		t.Run(name+" does nothing without results", func(t *testing.T) {
			t.Parallel()

			screen := searchingIn(t, nil)

			screen.Press(pressed)

			assert.Empty(t, screen.Outcomes())
			assert.True(t, screen.IsOpen())
		})
	}
}
