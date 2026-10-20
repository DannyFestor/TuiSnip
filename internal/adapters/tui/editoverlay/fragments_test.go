package editoverlay_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	twoTabStrip          = " main.go │ YAML "
	removeFragmentTitle  = "Remove Fragment"
	removeYAMLQuestion   = `Remove Fragment "YAML"? [y/N]`
	lastFragmentRefusal  = "A Snippet needs a Fragment; delete the Snippet instead"
	fragmentTitleTooLong = "Fragment title is longer than 200 characters"
)

func TestEditing_FragmentTabs(t *testing.T) {
	t.Parallel()

	t.Run("shows the tab strip and the first Fragment's title, Language and content", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))

		assert.Contains(t, screen.Screen(), twoTabStrip)
		assert.Contains(t, screen.Screen(), "Fragment title main.go")
		assert.Contains(t, screen.Screen(), "Language       Go")
		assert.Contains(t, screen.Screen(), "package main")
	})

	t.Run("shows no strip for one untitled Fragment, but the Fragment title row", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))

		assert.NotContains(t, screen.Screen(), " Go │")
		assert.Contains(t, screen.Screen(), "Fragment title")
	})

	t.Run("next Fragment shows the following Fragment's rows", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment())

		assert.Contains(t, screen.Screen(), "Language       YAML")
		assert.Contains(t, screen.Screen(), "on: push")
		assert.NotContains(t, screen.Screen(), "package main")
	})

	t.Run("previous Fragment wraps to the last Fragment", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(previousFragment())

		assert.Contains(t, screen.Screen(), "on: push")
	})

	t.Run("a click on a tab shows that Fragment", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Click("YAML")

		assert.Contains(t, screen.Screen(), "on: push")
	})

	t.Run("a typed Fragment title relabels the tab", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment())
		screen.Press(toFragmentTitle()...)
		screen.Press(keypress.Typed("ci.yml")...)

		assert.Contains(t, screen.Screen(), " main.go │ ci.yml ")
	})

	t.Run("keeps each Fragment's edits while switching between them", func(t *testing.T) {
		t.Parallel()

		stored := deploySnippet(t)
		screen := editingStored(t, stored)
		screen.Press(append(enterContent(), keypress.Typed("// ")...)...)
		screen.Press(keypress.Special(tea.KeyEscape), nextFragment(), previousFragment(), save())

		want := updateInput(stored, "Deploy", "package main\n// ")
		assert.Equal(t, []outcome.Outcome{outcome.UpdateRequested{Input: want}}, screen.Outcomes())
	})
}

func TestEditing_FragmentChanges(t *testing.T) {
	t.Parallel()

	t.Run(
		"adds a Fragment after the Shown Fragment, in the Folder's Default Language, and shows it",
		func(t *testing.T) {
			t.Parallel()

			stored := deploySnippet(t)
			screen := editingStoredIn(t, stored, language(t, "Bash"))
			screen.Press(addFragment())

			assert.Contains(t, screen.Screen(), " main.go │ Bash │ YAML ")
			assert.Contains(t, screen.Screen(), "Language       Bash")
			assert.Contains(t, screen.Screen(), unsavedTitle)

			screen.Press(save())

			fragments := savedFragments(t, screen)
			require.Len(t, fragments, 3)
			assert.Equal(t, snippet.FragmentInput{ID: domain.FragmentID{}, Language: "Bash"}, fragments[1])
		},
	)

	t.Run("moves the Shown Fragment right and left", func(t *testing.T) {
		t.Parallel()

		stored := deploySnippet(t)
		screen := editingStored(t, stored)
		screen.Press(moveFragmentRight())

		assert.Contains(t, screen.Screen(), " YAML │ main.go ")

		screen.Press(moveFragmentLeft(), moveFragmentRight(), save())

		assert.Equal(t, []string{"YAML", "Go"}, savedLanguages(t, screen))
	})

	t.Run("moves no Fragment past either end", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(moveFragmentLeft())

		assert.Contains(t, screen.Screen(), twoTabStrip)
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("asks before removing the Shown Fragment", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment(), removeFragment())

		assert.Contains(t, screen.Screen(), removeFragmentTitle)
		assert.Contains(t, screen.Screen(), removeYAMLQuestion)
	})

	t.Run("removes the Shown Fragment once confirmed", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment(), removeFragment(), keypress.Letter('y'))

		assert.NotContains(t, screen.Screen(), twoTabStrip)
		assert.Contains(t, screen.Screen(), "package main")

		screen.Press(save())

		assert.Equal(t, []string{"Go"}, savedLanguages(t, screen))
	})

	t.Run("keeps the Fragment when the removal is declined", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment(), removeFragment(), keypress.Letter('n'))

		assert.Contains(t, screen.Screen(), twoTabStrip)
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("refuses to remove the last Fragment", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))
		screen.Press(removeFragment())

		assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: lastFragmentRefusal}}, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), removeFragmentTitle)
	})

	t.Run("saves a retitled Fragment", func(t *testing.T) {
		t.Parallel()

		stored := deploySnippet(t)
		screen := editingStored(t, stored)
		screen.Press(toFragmentTitle()...)
		screen.Press(keypress.Special(tea.KeyEnd), keypress.Letter('x'), save())

		assert.Equal(t, "main.gox", savedFragments(t, screen)[0].Title)
	})

	t.Run("asks before cancelling Fragment changes", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(addFragment(), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), discardQuestion)
	})

	t.Run("adds a Fragment to a new Snippet in the destination's Language", func(t *testing.T) {
		t.Parallel()

		screen := editingIn(t, destinationInGo(t))
		screen.Press(addFragment(), save())

		requested, ok := screen.Outcomes()[0].(outcome.SaveRequested)
		require.True(t, ok)
		assert.Equal(t, []snippet.FragmentInput{{Language: "Go"}, {Language: "Go"}}, requested.Input.Fragments)
	})
}

func TestEditing_FragmentExternalEdit(t *testing.T) {
	t.Parallel()

	t.Run("opens the Shown Fragment in the external editor", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, deploySnippet(t))
		screen.Press(nextFragment(), externalEditor())

		assert.Equal(t, []outcome.Outcome{externalEditAsked("on: push\n", language(t, "YAML"))}, screen.Outcomes())
	})

	t.Run("puts the edited content into the Shown Fragment", func(t *testing.T) {
		t.Parallel()

		stored := deploySnippet(t)
		screen := editingStored(t, stored)
		screen.Press(nextFragment())
		screen.Offer(contentEdited("on: pull_request\n"))
		screen.Press(save())

		fragments := savedFragments(t, screen)
		assert.Equal(t, "package main\n", fragments[0].Content)
		assert.Equal(t, "on: pull_request\n", fragments[1].Content)
	})
}

func TestEditing_FragmentRefusedSave(t *testing.T) {
	t.Parallel()

	titleTooLong := domain.OnField(domain.FieldFragmentTitle, value.ErrFragmentTitleTooLong)

	t.Run("shows the Fragment with the invalid field and marks it", func(t *testing.T) {
		t.Parallel()

		screen := refusedSaving(t, domain.InFragment(1, titleTooLong))
		unmarked := editingStored(t, deploySnippet(t))
		unmarked.Press(nextFragment())

		assert.Contains(t, screen.Screen(), "on: push")
		assert.NotEqual(t, styledRow(t, unmarked, "Fragment title"), styledRow(t, screen, "Fragment title"))
		assert.Contains(t, screen.Outcomes(), outcome.Outcome(outcome.NoticeShown{Text: fragmentTitleTooLong}))
	})

	t.Run("drops the Fragment's marks when another Fragment is shown", func(t *testing.T) {
		t.Parallel()

		screen := refusedSaving(t, domain.InFragment(1, titleTooLong))
		screen.Press(previousFragment())

		unmarked := editingStored(t, deploySnippet(t))
		assert.Equal(t, styledRow(t, unmarked, "Fragment title"), styledRow(t, screen, "Fragment title"))
	})

	t.Run("marks only the shown Fragment's fields when several Fragments are invalid", func(t *testing.T) {
		t.Parallel()

		screen := refusedSaving(t, errors.Join(
			domain.InFragment(0, domain.OnField(domain.FieldContent, value.ErrContentTooLong)),
			domain.InFragment(1, titleTooLong),
		))

		unmarked := editingStored(t, deploySnippet(t))
		assert.Contains(t, screen.Screen(), "package main")
		assert.NotEqual(t, styledRow(t, unmarked, "Content"), styledRow(t, screen, "Content"))
		assert.Equal(t, styledRow(t, unmarked, "Fragment title"), styledRow(t, screen, "Fragment title"))
	})
}

func refusedSaving(t *testing.T, refusal error) *overlaytest.Driver {
	t.Helper()

	screen := editingStored(t, deploySnippet(t))
	screen.Press(save())
	screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: fmt.Errorf("snippet.Update: %w", refusal)})

	return screen
}

func styledRow(t *testing.T, screen *overlaytest.Driver, label string) string {
	t.Helper()

	for row := range strings.SplitSeq(screen.StyledScreen(), "\n") {
		if strings.Contains(row, label+" ") {
			return row
		}
	}

	require.Fail(t, "no row labelled "+label)

	return ""
}

func deploySnippet(t *testing.T) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return testkit.Snippet(t, testkit.SnippetSpec{
		ID:    ids.NewSnippetID(),
		Title: "Deploy",
		Fragments: []testkit.FragmentSpec{
			{ID: ids.NewFragmentID(), Title: "main.go", Language: "Go", Content: "package main\n"},
			{ID: ids.NewFragmentID(), Language: "YAML", Content: "on: push\n"},
		},
	})
}

func editingStoredIn(t *testing.T, stored domain.Snippet, defaultLanguage value.Language) *overlaytest.Driver {
	t.Helper()

	opened, _ := editoverlay.Editing(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		editoverlay.Options{},
		editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed(), DefaultLanguage: defaultLanguage},
	)

	return overlaytest.Open(t, screenSize(), opened)
}

func savedFragments(t *testing.T, screen *overlaytest.Driver) []snippet.FragmentInput {
	t.Helper()

	require.NotEmpty(t, screen.Outcomes())
	requested, ok := screen.Outcomes()[len(screen.Outcomes())-1].(outcome.UpdateRequested)
	require.True(t, ok)

	return requested.Input.Fragments
}

func savedLanguages(t *testing.T, screen *overlaytest.Driver) []string {
	t.Helper()

	fragments := savedFragments(t, screen)
	languages := make([]string, 0, len(fragments))

	for _, fragment := range fragments {
		languages = append(languages, fragment.Language)
	}

	return languages
}

func nextFragment() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModAlt}
}

func previousFragment() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt}
}

func moveFragmentRight() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModAlt | tea.ModShift}
}

func moveFragmentLeft() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt | tea.ModShift}
}

func addFragment() tea.KeyPressMsg {
	return keypress.Ctrl('n')
}

func removeFragment() tea.KeyPressMsg {
	return keypress.Ctrl('x')
}
