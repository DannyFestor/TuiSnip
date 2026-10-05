package mainscreen_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
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
			name:  "enter drills from Tags to the Snippet list",
			keys:  []tea.KeyPressMsg{keypress.Letter('2'), keypress.Special(tea.KeyEnter)},
			title: snippetListTitle,
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

func TestScreen_browse(t *testing.T) {
	t.Parallel()

	t.Run("shows the Folder tree", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		assert.Contains(t, screen.Screen(), "▾ go")
		assert.Contains(t, screen.Screen(), "    testing")
	})

	t.Run("moving the cursor onto a Folder selects it", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Press(keypress.Typed("jj")...)

		assert.Equal(t, []outcome.Outcome{
			outcome.FolderSelected{ID: sample.Docker.ID()},
			outcome.FolderSelected{ID: sample.Go.ID()},
		}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 Root / go · by title")
	})

	t.Run("lists the Snippets loaded for the selected Folder", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		filed := filedIn(t, sample.Go.ID())

		screen.Press(keypress.Typed("jj")...)

		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Go.ID()),
				Snippets:  []domain.Snippet{filed},
			},
		)

		assert.Contains(t, screen.Screen(), filedTitle)
		assert.NotContains(t, screen.Screen(), secondTitle)
	})

	t.Run("starts a newly selected Folder's list on its first Snippet", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("3j1j")...)

		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Docker.ID()),
				Snippets:  sampleSnippets(t),
			},
		)

		assert.Contains(t, screen.Screen(), firstDescription)
	})

	t.Run("ignores Snippets loaded for a Folder no longer selected", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Letter('j'))

		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Go.ID()),
				Snippets:  []domain.Snippet{filedIn(t, sample.Go.ID())},
			},
		)

		assert.NotContains(t, screen.Screen(), filedTitle)
	})

	t.Run("moving focus keeps the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Letter('j'))

		screen.Press(keypress.Special(tea.KeyTab), keypress.Letter('j'), keypress.Special(tea.KeyTab))

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 Root / docker · by title")
	})

	t.Run("shows the Snippet's Folder path in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("G")...)

		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Tests.ID()),
				Snippets:  []domain.Snippet{filedIn(t, sample.Tests.ID())},
			},
		)

		assert.Contains(t, screen.Screen(), "Root / go / testing · Go")
	})

	t.Run("shows each Search result's Folder path", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("G")...)
		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Tests.ID()),
				Snippets:  []domain.Snippet{filedIn(t, sample.Tests.ID())},
			},
		)

		screen.Press(keypress.Letter('/'))

		assert.Regexp(t, filedTitle+` +go / testing`, screen.Screen())
	})
}

func TestScreen_browseTags(t *testing.T) {
	t.Parallel()

	t.Run("shows the Tags with their Snippet counts", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)

		assert.Regexp(t, `# docker +3`, screen.Screen())
		assert.Regexp(t, `# go +4`, screen.Screen())
	})

	t.Run("moving the cursor onto a Tag selects it", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)

		screen.Press(keypress.Letter('2'), keypress.Letter('j'))

		assert.Equal(t, []outcome.Outcome{outcome.TagSelected{ID: tags[1].Tag.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 # go · by title")
	})

	t.Run("enter makes the Tag under the cursor the Browse selection and focuses the Snippet list", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)

		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagSelected{ID: tags[0].Tag.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), strings.ToUpper("3 # docker · by title"))
	})

	t.Run("enter on the Tag that already holds the selection selects nothing again", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		assert.Len(t, screen.Outcomes(), 1)
	})

	t.Run("enter with no Tags keeps the Folder selection", func(t *testing.T) {
		t.Parallel()

		screen := browsingWithTags(t)

		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		assert.Empty(t, screen.Outcomes())
		assert.Contains(t, screen.Screen(), strings.ToUpper(snippetListTitle))
	})

	t.Run("enter on Folders takes the selection back from a Tag", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Letter('j'))

		screen.Press(keypress.Letter('1'), keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{
			outcome.TagSelected{ID: tags[1].Tag.ID()},
			outcome.FolderSelected{ID: domain.FolderID{}},
		}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), strings.ToUpper(snippetListTitle))
	})

	t.Run("lists the Snippets loaded for the selected Tag and ignores those of the Folder", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		carrying := filedIn(t, domain.FolderID{})

		screen.Press(keypress.Letter('2'), keypress.Letter('j'))
		screen.Send(mainscreen.SnippetsLoaded{Selection: browseselection.InFolder(domain.FolderID{})})
		screen.Send(mainscreen.SnippetsLoaded{
			Selection: browseselection.WithTag(tags[1].Tag.ID()),
			Snippets:  []domain.Snippet{carrying},
		})

		assert.Contains(t, screen.Screen(), filedTitle)
		assert.NotContains(t, screen.Screen(), secondTitle)
	})

	t.Run("esc from the Snippet list returns to Tags while a Tag holds the selection", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), strings.ToUpper(tagPaneTitle))
	})

	t.Run("enter on Folders takes the selection back once the Tags are gone", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))
		screen.Send(mainscreen.TagsLoaded{Tags: nil})

		screen.Press(keypress.Letter('1'), keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), strings.ToUpper(folderPaneTitle))
	})

	t.Run("s asks for the selected Tag", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		screen.Press(keypress.Letter('s'))

		assert.Contains(t, screen.Outcomes(), outcome.SortCycleAsked{
			Selection: browseselection.WithTag(tags[0].Tag.ID()),
			Selecting: sampleSnippets(t)[0].ID(),
		})
	})

	t.Run("keeps a Tag tall in the left column while it holds the selection", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)

		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		assert.Equal(t, 13, lineIndexOf(screen, tagPaneTitle))
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

	t.Run("z fills the screen with the focused Pane", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('3'), keypress.Letter('z'))

		assert.Equal(t, []int{wide().Width}, columnWidths(screen))
		assert.Contains(t, screen.Screen(), snippetListTitle)
		assert.NotContains(t, screen.Screen(), snippetPaneTitle)
	})

	t.Run("zooming Folders shows Folders without Tags", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('z'))

		assert.Contains(t, screen.Screen(), folderPaneTitle)
		assert.NotContains(t, screen.Screen(), tagPaneTitle)
	})

	t.Run("z again restores the layout", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('z'), keypress.Letter('z'))

		assert.Equal(t, []int{34, 36, 50}, columnWidths(screen))
	})

	t.Run("moving focus while zoomed shows the newly focused Pane zoomed", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('z'), keypress.Letter('4'))

		assert.Equal(t, []int{wide().Width}, columnWidths(screen))
		assert.Contains(t, screen.Screen(), snippetPaneTitle)
		assert.NotContains(t, screen.Screen(), folderPaneTitle)
	})

	t.Run("zooming Tags shows Tags without Folders", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)

		screen.Press(keypress.Letter('2'), keypress.Letter('z'))

		assert.Equal(t, []int{wide().Width}, columnWidths(screen))
		assert.Contains(t, screen.Screen(), strings.ToUpper(tagPaneTitle))
		assert.Contains(t, screen.Screen(), "# docker")
		assert.NotContains(t, screen.Screen(), folderPaneTitle)
	})

	zoomedEnters := []struct {
		name      string
		from      rune
		leftTitle string
	}{
		{name: "enter on Folders while zoomed shows the Snippet list zoomed", from: '1', leftTitle: folderPaneTitle},
		{name: "enter on Tags while zoomed shows the Snippet list zoomed", from: '2', leftTitle: tagPaneTitle},
	}

	for _, tt := range zoomedEnters {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen, _ := browsingTags(t)

			screen.Press(keypress.Letter(tt.from), keypress.Letter('z'), keypress.Special(tea.KeyEnter))

			assert.Equal(t, []int{wide().Width}, columnWidths(screen))
			assert.Contains(t, screen.Screen(), " BY TITLE")
			assert.NotContains(t, strings.ToUpper(screen.Screen()), strings.ToUpper(tt.leftTitle))
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

	t.Run("cuts a status text too long to fit beside help", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('3'))

		room := wide().Width - ansi.StringWidth(helpHint) - 1
		screen.Send(mainscreen.StatusShown{Text: strings.Repeat("x", room)})

		assert.Equal(t, " "+strings.Repeat("x", room-2)+"… "+helpHint, statusLine(screen))
	})

	t.Run("keeps every hint that fits beside the status text", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('3'))

		screen.Send(mainscreen.StatusShown{Text: statusLeaving(ansi.StringWidth(listHint))})

		assert.True(t, strings.HasSuffix(statusLine(screen), " "+listHint))
	})

	t.Run("drops hints from the right but keeps help", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('3'))

		screen.Send(mainscreen.StatusShown{Text: statusLeaving(ansi.StringWidth(listHint) - 1)})

		assert.True(t, strings.HasSuffix(statusLine(screen), " y Copy · e edit · n new · s sort · z zoom · "+helpHint))
	})

	t.Run("keeps only help when nothing else fits", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('3'))

		screen.Send(mainscreen.StatusShown{Text: statusLeaving(ansi.StringWidth(helpHint))})

		assert.True(t, strings.HasSuffix(statusLine(screen), "x "+helpHint))
	})

	t.Run("drops an Overlay's hints from the right", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, minimum(), sampleSnippets(t)...)
		screen.Press(keypress.Letter('/'))

		screen.Send(mainscreen.StatusShown{Text: statusLeaving(ansi.StringWidth(searchHint) - 1)})

		assert.True(t, strings.HasSuffix(statusLine(screen), " down move · enter reveal · ctrl+y Copy"))
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

	t.Run("keeps the focused Pane's hints while zoomed", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('3'), keypress.Letter('z'))

		assert.True(t, strings.HasSuffix(statusLine(screen), " "+listHint))
	})

	t.Run("shows an open Overlay's hints instead of the too-small hint", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, narrow(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('/'))

		assert.NotContains(t, statusLine(screen), tooSmallHint)
		assert.Contains(t, statusLine(screen), "down move")
	})
}

func TestScreen_restyle(t *testing.T) {
	t.Parallel()

	t.Run("draws every Pane in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)
		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Send(look.Restyled{Styles: light})

		assert.Equal(t, showingStyled(t, wide(), light, sampleSnippets(t)...).StyledScreen(), screen.StyledScreen())
	})

	t.Run("draws the zoomed Pane in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)
		screen := showing(t, wide(), sampleSnippets(t)...)
		styledFromStart := showingStyled(t, wide(), light, sampleSnippets(t)...)

		screen.Press(keypress.Typed("3z")...)
		screen.Send(look.Restyled{Styles: light})
		styledFromStart.Press(keypress.Typed("3z")...)

		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})

	openings := []struct {
		typed string
		title string
	}{
		{typed: "n", title: "╭ Editing "},
		{typed: "/", title: "╭ Search "},
		{typed: "?", title: "╭ Help "},
		{typed: "3e", title: "╭ Editing "},
	}
	for _, opening := range openings {
		t.Run("opens the Overlay on "+opening.typed+" in the new Styles", func(t *testing.T) {
			t.Parallel()

			light := look.NewStyles(look.SchemeLight)
			screen := showing(t, wide(), sampleSnippets(t)...)
			screen.Send(look.Restyled{Styles: light})
			styledFromStart := showingStyled(t, wide(), light, sampleSnippets(t)...)

			screen.Press(keypress.Typed(opening.typed)...)
			styledFromStart.Press(keypress.Typed(opening.typed)...)

			assert.Contains(t, screen.Screen(), opening.title)
			assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
		})
	}
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

	t.Run("e in the Snippet list edits the selected Snippet", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("3e")...)

		assert.Contains(t, screen.Screen(), "› Title       Graceful HTTP shutdown")
		assert.Equal(t, "ctrl+s save · esc cancel · down field", screen.Hints())
	})

	t.Run("e in the Snippet pane edits the shown Snippet", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("3j4e")...)

		assert.Contains(t, screen.Screen(), "› Title       "+secondTitle)
	})

	t.Run("e edits nothing outside the Snippet list and the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('e'))

		assert.NotContains(t, screen.Screen(), "Editing")
	})

	t.Run("e edits nothing in an empty Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Typed("3e4e")...)

		assert.NotContains(t, screen.Screen(), "Editing")
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

	t.Run("keys type into a Folder name instead of acting", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('N'))
		screen.Press(keypress.Typed("qn/2")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.FolderCreateRequested{
			Input: folder.CreateInput{Name: "qn/2", ParentID: domain.FolderID{}},
		}}, screen.Outcomes())
	})

	t.Run("a paste goes into a Folder name", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('N'))
		screen.Send(tea.PasteMsg{Content: "docker"})
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.FolderCreateRequested{
			Input: folder.CreateInput{Name: "docker", ParentID: domain.FolderID{}},
		}}, screen.Outcomes())
	})

	t.Run("hints save and cancel while a Folder name is typed", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Letter('N'))

		assert.Equal(t, "enter save · esc cancel", screen.Hints())
	})

	t.Run("makes a created Folder the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		created := testkit.Folder(t, testkit.FolderSpec{ID: testkit.NewSequentialIDs().NewFolderID(), Name: "awk"})
		grown := sample.Tree
		grown.Folders = append([]browse.FolderNode{{Folder: created, SnippetCount: 0, Children: nil}}, grown.Folders...)

		screen.Send(mainscreen.TreeChanged{Tree: grown, Selecting: created.ID()})

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: created.ID()}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 Root / awk · by title")
	})
}

const (
	nothingTyped = ""
	capturedText = "docker ps\n"
)

func TestScreen_newSnippet(t *testing.T) {
	t.Parallel()

	for _, place := range browseSelections() {
		t.Run("n files the Snippet for "+place.name, func(t *testing.T) {
			t.Parallel()

			screen, want := place.browse(t)

			screen.Press(keypress.Typed("nx")...)
			screen.Press(keypress.Ctrl('s'))

			assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: want(nothingTyped)}}, lastOutcome(screen))
		})

		t.Run("captured content is filed for "+place.name, func(t *testing.T) {
			t.Parallel()

			screen, want := place.browse(t)

			screen.Press(keypress.Letter('p'))
			screen.Send(mainscreen.Captured{Content: capturedText})
			screen.Press(keypress.Letter('x'), keypress.Ctrl('s'))

			assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: want(capturedText)}}, lastOutcome(screen))
		})
	}

	t.Run("p asks for Capture without opening anything", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('p'))

		assert.Equal(t, []outcome.Outcome{outcome.CaptureAsked{}}, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), "Editing")
	})

	t.Run("captured content opens the edit overlay on Title as an unsaved change", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Send(mainscreen.Captured{Content: capturedText})

		assert.Contains(t, screen.Screen(), "Editing •")
		assert.Contains(t, screen.Screen(), "› Title")
		assert.Contains(t, screen.Screen(), strings.TrimSpace(capturedText))
	})

	t.Run("captured content over 10,000 lines is refused on the status line and opens nothing", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Send(mainscreen.Captured{Content: strings.Repeat("line\n", 10_000)})

		assert.Contains(t, statusLine(screen), "Clipboard is longer than 10,000 lines; use ctrl+e to edit in $EDITOR")
		assert.NotContains(t, screen.Screen(), "Editing")
	})

	t.Run("captured content with tabs is highlighted in the selected Folder's Default Language", func(t *testing.T) {
		t.Parallel()

		inGo := capturedInFolderOf(t, "Go")
		inPlainText := capturedInFolderOf(t, value.PlainText().String())

		assert.Contains(t, inGo.Screen(), "Contains tabs")
		assert.Equal(t, inPlainText.Screen(), inGo.Screen())
		assert.NotEqual(t, inPlainText.StyledScreen(), inGo.StyledScreen())
	})
}

func capturedInFolderOf(t *testing.T, defaultLanguage string) *overlaytest.Driver {
	t.Helper()

	filing := testkit.Folder(t, testkit.FolderSpec{Name: "go", DefaultLanguage: defaultLanguage})
	screen := showing(t, wide())
	screen.Send(mainscreen.TreeLoaded{Tree: browse.Tree{
		RootSnippetCount: 0,
		Folders:          []browse.FolderNode{{Folder: filing, SnippetCount: 0, Children: nil}},
	}})
	screen.Press(keypress.Letter('j'))
	screen.Send(mainscreen.Captured{Content: "if x {\n\treturn\n}\n"})

	return screen
}

type createdFor func(content string) snippet.CreateInput

type browseSelection struct {
	name   string
	browse func(t *testing.T) (*overlaytest.Driver, createdFor)
}

func browseSelections() []browseSelection {
	return []browseSelection{
		{name: "the Root at the Root", browse: browsingRoot},
		{name: "a Folder in that Folder", browse: browsingFolder},
		{name: "a Tag at the Root carrying the Tag", browse: browsingTag},
	}
}

func browsingRoot(t *testing.T) (*overlaytest.Driver, createdFor) {
	t.Helper()

	return showing(t, wide()), createdIn(domain.FolderID{})
}

func browsingFolder(t *testing.T) (*overlaytest.Driver, createdFor) {
	t.Helper()

	screen, sample := browsing(t)
	screen.Press(keypress.Letter('j'))

	return screen, createdIn(sample.Docker.ID())
}

func browsingTag(t *testing.T) (*overlaytest.Driver, createdFor) {
	t.Helper()

	screen, tags := browsingTags(t)
	screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

	return screen, createdIn(domain.FolderID{}, tags[0].Tag)
}

func createdIn(folderID domain.FolderID, tags ...domain.Tag) createdFor {
	return func(content string) snippet.CreateInput {
		return snippet.CreateInput{Title: "x", Description: "", Content: content, FolderID: folderID, Tags: tags}
	}
}

func lastOutcome(screen *overlaytest.Driver) []outcome.Outcome {
	outcomes := screen.Outcomes()

	return outcomes[len(outcomes)-1:]
}

func TestScreen_sortOrder(t *testing.T) {
	t.Parallel()

	t.Run("s in the Snippet list asks to cycle the sort order, keeping the selected Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("3js")...)

		assert.Equal(t, []outcome.Outcome{outcome.SortCycleAsked{
			Selection: browseselection.Selection{},
			Selecting: snippets[1].ID(),
		}}, screen.Outcomes())
	})

	t.Run("s asks for the selected Folder", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("jj")...)
		screen.Send(mainscreen.SnippetsLoaded{
			Selection: browseselection.InFolder(sample.Go.ID()),
			Snippets:  []domain.Snippet{filedIn(t, sample.Go.ID())},
			Order:     domain.SortOrderTitle,
		})

		screen.Press(keypress.Typed("3s")...)

		assert.Contains(t, screen.Outcomes(), outcome.SortCycleAsked{
			Selection: browseselection.InFolder(sample.Go.ID()),
			Selecting: filedIn(t, sample.Go.ID()).ID(),
		})
	})

	t.Run("s does nothing outside the Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("s2s4s")...)

		assert.Empty(t, screen.Outcomes())
	})

	t.Run("starts with the order it was opened with", func(t *testing.T) {
		t.Parallel()

		screen := opened(t, wide(), look.NewStyles(look.SchemeDark), domain.SortOrderCreated)

		assert.Contains(t, screen.Screen(), "3 Root · by creation date")
	})

	titles := []struct {
		order domain.SortOrder
		title string
	}{
		{order: domain.SortOrderTitle, title: "3 Root · by title"},
		{order: domain.SortOrderUpdated, title: "3 Root · by last updated"},
		{order: domain.SortOrderCreated, title: "3 Root · by creation date"},
	}
	for _, tt := range titles {
		t.Run("titles the Snippet list with the order it was loaded in: "+tt.title, func(t *testing.T) {
			t.Parallel()

			snippets := sampleSnippets(t)
			screen := showing(t, wide(), snippets...)

			screen.Send(mainscreen.SnippetsLoaded{Snippets: snippets, Order: tt.order})

			assert.Contains(t, screen.Screen(), tt.title)
		})
	}

	t.Run("keeps the title of the shown list when Snippets load for another Folder", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Go.ID()),
				Order:     domain.SortOrderCreated,
			},
		)

		assert.Contains(t, screen.Screen(), "3 Root · by title")
	})

	t.Run("rejects an unknown order, keeping the title and listing the Snippets", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide())

		screen.Send(mainscreen.SnippetsLoaded{Snippets: snippets, Order: domain.SortOrder("language")})

		assert.Contains(t, screen.Screen(), "3 Root · by title")
		assert.Contains(t, screen.Screen(), secondTitle)
		require.Len(t, screen.Outcomes(), 1)
		rejected, ok := screen.Outcomes()[0].(outcome.SortOrderRejected)
		require.True(t, ok)
		assert.ErrorIs(t, rejected.Err, domain.ErrInvalidSortOrder)
	})
}

func TestScreen_UpdateFolderDelete(t *testing.T) {
	t.Parallel()

	t.Run("d passes on the ask to delete the Folder under the cursor", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Press(keypress.Letter('j'), keypress.Letter('d'))

		assert.Contains(t, screen.Outcomes(), outcome.FolderDeleteAsked{ID: sample.Docker.ID()})
	})

	t.Run("asks to confirm with the counts from the preview", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)

		screen.Send(mainscreen.FolderDeletePreviewed{
			Preview: folder.DeletePreview{Folder: sample.Docker, SubfolderCount: 1, SnippetCount: 15},
		})

		assert.Contains(t, screen.Screen(), "Delete Folder")
		assert.Contains(
			t,
			screen.Screen(),
			`Permanently delete Folder "docker", 1 subfolder, and 15 Snippets? This cannot be undone. [y/N]`,
		)
	})

	t.Run("yes asks to delete the Folder, remembering its parent", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Send(mainscreen.FolderDeletePreviewed{
			Preview: folder.DeletePreview{Folder: sample.Tests, SubfolderCount: 0, SnippetCount: 1},
		})

		screen.Press(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.FolderDeleteRequested{
			Input:    folder.DeleteInput{FolderID: sample.Tests.ID()},
			ParentID: sample.Go.ID(),
		}}, screen.Outcomes())
	})

	t.Run("no closes the confirmation without deleting", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Send(mainscreen.FolderDeletePreviewed{
			Preview: folder.DeletePreview{Folder: sample.Docker, SubfolderCount: 0, SnippetCount: 0},
		})

		screen.Press(keypress.Letter('n'))

		assert.Empty(t, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), "Delete Folder")
	})

	t.Run("makes the deleted Folder's parent the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Letter('j'), keypress.Letter('j'), keypress.Letter('j'))

		shrunk := sample.Tree
		shrunk.Folders = []browse.FolderNode{sample.Tree.Folders[0], sample.Tree.Folders[1]}
		shrunk.Folders[1].Children = nil

		screen.Send(mainscreen.TreeChanged{Tree: shrunk, Selecting: sample.Go.ID()})

		assert.Contains(t, screen.Outcomes(), outcome.FolderSelected{ID: sample.Go.ID()})
		assert.Contains(t, screen.Screen(), "3 Root / go · by title")
	})
}

func TestScreen_folderDeleteQuestion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subfolders int
		snippets   int
		want       string
	}{
		{name: "pluralises zero counts", subfolders: 0, snippets: 0, want: "0 subfolders, and 0 Snippets?"},
		{name: "keeps single counts singular", subfolders: 1, snippets: 1, want: "1 subfolder, and 1 Snippet?"},
		{name: "pluralises larger counts", subfolders: 2, snippets: 15, want: "2 subfolders, and 15 Snippets?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen, sample := browsing(t)

			screen.Send(mainscreen.FolderDeletePreviewed{
				Preview: folder.DeletePreview{
					Folder:         sample.Docker,
					SubfolderCount: tt.subfolders,
					SnippetCount:   tt.snippets,
				},
			})

			assert.Contains(t, screen.Screen(), `Permanently delete Folder "docker", `+tt.want)
		})
	}
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

	t.Run("makes a revealed Snippet's Folder the Browse selection", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Letter('2'))

		filed := filedIn(t, sample.Go.ID())

		screen.Offer(outcome.SnippetRevealed{ID: filed.ID(), FolderID: sample.Go.ID()})
		screen.Send(
			mainscreen.SnippetsLoaded{
				Selection: browseselection.InFolder(sample.Go.ID()),
				Snippets:  []domain.Snippet{filed},
				Selecting: filed.ID(),
				Order:     domain.SortOrderTitle,
			},
		)

		assert.Contains(t, screen.Screen(), "3 Root / go · by title")
		assert.Contains(t, screen.Screen(), filedTitle)
		assert.Equal(t, 26, lineIndexOf(screen, tagPaneTitle), "Folders, holding the Browse selection, is tall")
	})

	t.Run("expands the collapsed Folders hiding a revealed Snippet's Folder", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("jj ")...)

		revealed := outcome.SnippetRevealed{ID: filedIn(t, sample.Tests.ID()).ID(), FolderID: sample.Tests.ID()}
		screen.Offer(revealed)

		assert.Equal(t, []outcome.Outcome{
			outcome.CollapsedFoldersChanged{IDs: []domain.FolderID{}},
			revealed,
		}, screen.Outcomes()[len(screen.Outcomes())-2:])
		assert.Contains(t, screen.Screen(), "3 Root / go / testing")
	})

	t.Run("keeps the Browse selection for a saved Snippet and passes it on", func(t *testing.T) {
		t.Parallel()

		screen, tags := browsingTags(t)
		screen.Press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))

		saved := outcome.SnippetSaved{
			ID:        filedIn(t, domain.FolderID{}).ID(),
			Selection: browseselection.WithTag(tags[0].Tag.ID()),
		}

		screen.Offer(saved)

		assert.Equal(t, []outcome.Outcome{outcome.TagSelected{ID: tags[0].Tag.ID()}, saved}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), strings.ToUpper("3 # docker · by title"))
	})

	t.Run("keeps the Browse selection for a reloaded Snippet and passes it on", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Letter('j'))

		reloaded := outcome.SnippetReloaded{
			ID:        filedIn(t, domain.FolderID{}).ID(),
			Selection: browseselection.InFolder(sample.Docker.ID()),
		}

		screen.Offer(reloaded)

		assert.Equal(t, []outcome.Outcome{outcome.FolderSelected{ID: sample.Docker.ID()}, reloaded}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "3 Root / docker · by title")
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
	screen := newScreen(t, look.NewStyles(look.SchemeDark), domain.SortOrderTitle)

	want := append(keys.For(binding.ScopeGlobal).FullHelp(), folderpane.New(keys, nil).FullHelp()...)
	assert.Equal(t, want, screen.FullHelp())
}

func TestScreen_help(t *testing.T) {
	t.Parallel()

	t.Run("? opens help over the global Bindings and the focused Pane's", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('?'))

		assert.Contains(t, screen.Screen(), "╭ Help ")
		assert.Regexp(t, `q\s+quit`, screen.Screen())
		assert.Regexp(t, `N\s+new Folder`, screen.Screen())
	})

	t.Run("lists the Snippet list's Bindings when it has focus", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("3?")...)

		assert.Regexp(t, `s\s+sort`, screen.Screen())
		assert.NotContains(t, screen.Screen(), "new Folder")
	})

	t.Run("hints closing help", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('?'))

		assert.Equal(t, "? close · esc close", screen.Hints())
	})

	t.Run("opens below the minimum size", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, narrow())

		screen.Press(keypress.Letter('?'))

		assert.Contains(t, screen.Screen(), "╭ Help ")
	})

	t.Run("opens while a Pane is zoomed and lists zoom", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('z'), keypress.Letter('?'))

		assert.Contains(t, screen.Screen(), "╭ Help ")
		assert.Regexp(t, `z\s+zoom`, screen.Screen())
	})

	t.Run("opens from the Tags Pane over the global Bindings only", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsingTags(t)

		screen.Press(keypress.Letter('2'), keypress.Letter('?'))

		assert.Regexp(t, `q\s+quit`, screen.Screen())
		assert.NotContains(t, screen.Screen(), "new Folder")
		assert.Equal(t, "? close · esc close", screen.Hints())
	})

	t.Run("? types into the edit overlay's Content instead of opening help", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Letter('n'), keypress.Special(tea.KeyDown), keypress.Special(tea.KeyDown))
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Letter('?'), keypress.Ctrl('s'))

		assert.NotContains(t, screen.Screen(), "╭ Help ")
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{
			Input: snippet.CreateInput{Title: "", Description: "", Content: "?"},
		}}, screen.Outcomes())
	})

	t.Run("? types into the Search query instead of opening help", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Typed("/?")...)

		assert.Contains(t, screen.Screen(), "/ ?")
		assert.NotContains(t, screen.Screen(), "╭ Help ")
	})

	t.Run("? types into a Folder name instead of opening help", func(t *testing.T) {
		t.Parallel()

		screen, _ := browsing(t)

		screen.Press(keypress.Typed("N?")...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.FolderCreateRequested{
			Input: folder.CreateInput{Name: "?", ParentID: domain.FolderID{}},
		}}, screen.Outcomes())
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	_, err := mainscreen.New(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		time.UTC,
		mainscreen.Remembered{SortOrder: domain.SortOrder("language"), CollapsedFolders: nil},
	)

	require.ErrorIs(t, err, domain.ErrInvalidSortOrder)
	assert.ErrorContains(t, err, "mainscreen.New: ")
}
