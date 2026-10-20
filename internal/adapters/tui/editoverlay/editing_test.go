package editoverlay_test

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	tabbedContent = "if x {\n\treturn\n}\n"

	changedElsewhereTitle = "Changed elsewhere"
	reloadQuestion        = "This Snippet changed in another TuiSnip. Reload it and discard your changes? [y/N]"
)

var errChangedElsewhere = fmt.Errorf("snippet.Update: %w", domain.ErrConflict)

func TestEditing_View(t *testing.T) {
	t.Parallel()

	t.Run("opens filled from the Snippet on Title", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))

		assert.Contains(t, screen.Screen(), "› Title          Prune")
		assert.Contains(t, screen.Screen(), "Description    Reclaim")
		assert.Contains(t, screen.Screen(), "echo hi")
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("marks a change to the Snippet as unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Letter('x'))

		assert.Contains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("opens content with a tab editable and runs the tab to the next tab stop", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, tabbedContent))

		assert.Contains(t, screen.Screen(), contentEntryHint)
		assert.Contains(t, screen.Screen(), "    return")
		assert.NotContains(t, screen.Screen(), "read-only")
	})

	t.Run("draws the Content and the unsaved marker in the new Styles", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		light := look.NewStyles(look.SchemeLight)
		screen := editingStoredWith(t, keys, storedSnippet(t, tabbedContent))
		screen.Press(keypress.Letter('x'))

		screen.Send(look.Restyled{Styles: light})

		styledFromStart := editingStoredStyled(t, keys, light, storedSnippet(t, tabbedContent))
		styledFromStart.Press(keypress.Letter('x'))
		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})
}

func TestEditing_save(t *testing.T) {
	t.Parallel()

	t.Run("asks to update the Snippet guarded by its loaded time", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Press(keypress.Typed(" all")...)
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune all", "echo hi!")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("keeps content with tabs byte for byte when it is saved unchanged", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, tabbedContent)
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Press(keypress.Special(tea.KeyDown), keypress.Special(tea.KeyRight), keypress.Special(tea.KeyUp))
		screen.Press(save())

		assert.NotContains(t, screen.Screen(), unsavedTitle)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", tabbedContent)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("keeps the tabs of content edited around them", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, tabbedContent)
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Press(keypress.Special(tea.KeyUp), keypress.Special(tea.KeyUp), keypress.Special(tea.KeyRight))
		screen.Press(keypress.Typed("x")...)
		screen.Send(tea.PasteMsg{Content: "\ty"})
		screen.Press(save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", "if x {\n\tx\tyreturn\n}\n")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("closes and reloads the saved Snippet in the Browse selection it was opened from", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: stored, Err: nil})

		reported := outcome.SnippetReloaded{ID: stored.ID(), Selection: browsed()}
		assert.Contains(t, screen.Outcomes(), outcome.Outcome(reported))
		assert.False(t, screen.IsOpen())
	})
}

func TestEditing_language(t *testing.T) {
	t.Parallel()

	t.Run("shows the Fragment's Language", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))

		assert.Contains(t, screen.Screen(), "  Language       Go   (enter or ctrl+l to pick)")
	})

	t.Run("saves a picked Language", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)

		screen.Press(pickLanguage())
		screen.Press(keypress.Typed("bash")...)
		screen.Press(keypress.Special(tea.KeyEnter), save())

		want := updateInput(stored, "Prune", "echo hi")
		want.Fragments[0].Language = "Bash"
		assert.Equal(t, []outcome.Outcome{outcome.UpdateRequested{Input: want}}, screen.Outcomes())
	})

	t.Run("keeps a Language the curated list leaves out", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStoredOffering(t, []value.Language{language(t, "YAML")}, stored)

		screen.Press(keypress.Letter('!'), save())

		want := updateInput(stored, "Prune!", "echo hi")
		assert.Equal(t, "Go", want.Fragments[0].Language)
		assert.Equal(t, []outcome.Outcome{outcome.UpdateRequested{Input: want}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), "Language       Go")
	})

	t.Run("picking the same Language leaves nothing unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))

		screen.Press(pickLanguage(), keypress.Special(tea.KeyEnter))

		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})
}

func TestEditing_manyLines(t *testing.T) {
	t.Parallel()

	t.Run("opens stored content of 20,000 short lines editable and saves it byte for byte", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, twentyThousandShortLines())
		screen := editingStored(t, stored)

		assert.Contains(t, screen.Screen(), contentEntryHint)

		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", twentyThousandShortLines()+"!")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})
}

func TestEditing_editableContent(t *testing.T) {
	t.Parallel()

	t.Run("indents the cursor's line of the stored content", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "first\nsecond")
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Press(keypress.Special(tea.KeyTab), save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", "first\n    second")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run(
		"refuses a paste that makes the stored content larger than 256 KiB and leaves it unchanged",
		func(t *testing.T) {
			t.Parallel()

			screen := editingStored(t, storedSnippet(t, textOfSize(maxContentBytes)))
			screen.Press(enterContent()...)
			screen.Send(tea.PasteMsg{Content: "x"})

			assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: pasteTooLarge}}, screen.Outcomes())
			assert.NotContains(t, screen.Screen(), unsavedTitle)
		},
	)
}

func TestEditing_staleSave(t *testing.T) {
	t.Parallel()

	t.Run("offers to reload a Snippet changed elsewhere", func(t *testing.T) {
		t.Parallel()

		screen := refusedAsStale(t, testsettings.Default(t).Keys, storedSnippet(t, "echo hi"))

		assert.Contains(t, screen.Screen(), changedElsewhereTitle)
		assert.Contains(t, screen.Screen(), reloadQuestion)
		assert.Len(t, screen.Outcomes(), 1, "only the update request, no failure")
	})

	t.Run("closes and reloads the stored Snippet on yes", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := refusedAsStale(t, testsettings.Default(t).Keys, stored)
		screen.Press(keypress.Letter('y'))

		reloaded := outcome.SnippetReloaded{ID: stored.ID(), Selection: browsed()}
		assert.Contains(t, screen.Outcomes(), outcome.Outcome(reloaded))
		assert.False(t, screen.IsOpen())
	})

	t.Run("reloads on the configured yes key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeConfirm][binding.Yes] = []string{"o"}
		screen := refusedAsStale(t, keys, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Letter('o'))

		assert.False(t, screen.IsOpen())
	})

	t.Run("keeps the edits open on no", func(t *testing.T) {
		t.Parallel()

		screen := refusedAsStale(t, testsettings.Default(t).Keys, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Letter('n'))

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "Prunex")
		assert.NotContains(t, screen.Screen(), reloadQuestion)
		assert.Len(t, screen.Outcomes(), 1, "only the update request, no reload")
	})

	t.Run("asks to update again on the next save after no", func(t *testing.T) {
		t.Parallel()

		screen := refusedAsStale(t, testsettings.Default(t).Keys, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Letter('n'), save())

		assert.Len(t, screen.Outcomes(), 2)
		assert.IsType(t, outcome.UpdateRequested{}, screen.Outcomes()[1])
	})

	t.Run("asks in the Styles the overlay last got", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		light := look.NewStyles(look.SchemeLight)
		screen := editingStoredWith(t, keys, storedSnippet(t, "echo hi"))
		screen.Send(look.Restyled{Styles: light})

		refuseAsStale(screen)

		styledFromStart := refuseAsStale(editingStoredStyled(t, keys, light, storedSnippet(t, "echo hi")))
		assert.Contains(t, screen.Screen(), changedElsewhereTitle)
		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})

	t.Run("redraws the open question in the new Styles", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		light := look.NewStyles(look.SchemeLight)
		screen := refusedAsStale(t, keys, storedSnippet(t, "echo hi"))

		screen.Send(look.Restyled{Styles: light})

		styledFromStart := refuseAsStale(editingStoredStyled(t, keys, light, storedSnippet(t, "echo hi")))
		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})
}

func TestEditing_cancel(t *testing.T) {
	t.Parallel()

	t.Run("closes at once without changes", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, tabbedContent))
		screen.Press(keypress.Special(tea.KeyEscape))

		assert.False(t, screen.IsOpen())
	})

	t.Run("asks before discarding changes", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), discardQuestion)
	})
}

func refusedAsStale(t *testing.T, keys binding.Keys, stored domain.Snippet) *overlaytest.Driver {
	t.Helper()

	return refuseAsStale(editingStoredWith(t, keys, stored))
}

func refuseAsStale(screen *overlaytest.Driver) *overlaytest.Driver {
	screen.Press(keypress.Letter('x'), save())
	screen.Send(editoverlay.SaveFinished{Snippet: domain.Snippet{}, Err: errChangedElsewhere})

	return screen
}

func updateInput(stored domain.Snippet, title, firstContent string) snippet.UpdateInput {
	fragments := make([]snippet.FragmentInput, 0, len(stored.Fragments()))
	for _, fragment := range stored.Fragments() {
		fragments = append(fragments, snippet.FragmentInput{
			ID:       fragment.ID(),
			Title:    fragment.Title().String(),
			Language: fragment.Language().String(),
			Content:  fragment.Content().String(),
		})
	}

	fragments[0].Content = firstContent

	return snippet.UpdateInput{
		SnippetID:       stored.ID(),
		LoadedUpdatedAt: stored.UpdatedAt(),
		Title:           title,
		Description:     stored.Description().String(),
		Fragments:       fragments,
	}
}
