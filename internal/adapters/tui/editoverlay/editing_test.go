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
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	tabbedContent   = "if x {\n\treturn\n}\n"
	readOnlyNotice  = "Contains tabs: read-only here, edit with ctrl+e ($EDITOR)"
	unboundReadOnly = "Contains tabs: read-only here"

	changedElsewhereTitle = "Changed elsewhere"
	reloadQuestion        = "This Snippet changed in another TuiSnip. Reload it and discard your changes? [y/N]"
)

var errChangedElsewhere = fmt.Errorf("snippet.Update: %w", domain.ErrConflict)

func TestEditing_View(t *testing.T) {
	t.Parallel()

	t.Run("opens filled from the Snippet on Title", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))

		assert.Contains(t, screen.Screen(), "› Title       Prune")
		assert.Contains(t, screen.Screen(), "Description Reclaim")
		assert.Contains(t, screen.Screen(), "echo hi")
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("marks a change to the Snippet as unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))
		screen.Press(keypress.Letter('x'))

		assert.Contains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("shows content with a tab read-only and names the external editor key", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, tabbedContent))

		assert.Contains(t, screen.Screen(), readOnlyNotice)
		assert.Contains(t, screen.Screen(), "    return")
		assert.NotContains(t, screen.Screen(), contentEntryHint)
	})

	t.Run("names the configured external editor key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeContent][binding.OpenInEditor] = []string{"ctrl+o"}
		screen := editingStoredWith(t, keys, storedSnippet(t, tabbedContent))

		assert.Contains(t, screen.Screen(), "Contains tabs: read-only here, edit with ctrl+o ($EDITOR)")
	})

	t.Run("names no key when the external editor is unbound", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeContent][binding.OpenInEditor] = []string{}
		screen := editingStoredWith(t, keys, storedSnippet(t, tabbedContent))

		assert.Contains(t, screen.Screen(), unboundReadOnly)
		assert.NotContains(t, screen.Screen(), "edit with")
	})

	t.Run("draws the read-only Content and the unsaved marker in the new Styles", func(t *testing.T) {
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

	t.Run("keeps read-only content byte for byte", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, tabbedContent)
		screen := editingStored(t, stored)
		screen.Press(keypress.Typed(" all")...)
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune all", tabbedContent)}
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

func TestEditing_readOnlyContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sent tea.Msg
	}{
		{name: "indent leaves it unchanged", sent: keypress.Special(tea.KeyTab)},
		{name: "dedent leaves it unchanged", sent: shiftTab()},
		{name: "a paste leaves it unchanged", sent: tea.PasteMsg{Content: "pasted"}},
		{name: "a paste with tabs leaves it unchanged", sent: tea.PasteMsg{Content: "\tpasted"}},
		{name: "a paste over 10,000 lines leaves it unchanged", sent: tea.PasteMsg{Content: linesOf(10_001)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stored := storedSnippet(t, tabbedContent)
			screen := editingStored(t, stored)
			screen.Press(enterContent()...)
			screen.Send(tt.sent)
			screen.Press(save())

			want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", tabbedContent)}
			assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		})
	}
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

	t.Run("refuses a paste that makes the stored content longer than 10,000 lines", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, linesOf(5_000)))
		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: linesOf(5_002)})

		want := outcome.NoticeShown{
			Text: "Paste would make Content longer than 10,000 lines; use ctrl+e to edit in $EDITOR",
		}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("inserts a paste that makes the stored content exactly 10,000 lines", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, linesOf(5_000))
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Send(tea.PasteMsg{Content: linesOf(5_001)})
		screen.Press(save())

		content := linesOf(4_999) + "\nlineline\n" + linesOf(5_000)
		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", content)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})
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

func updateInput(stored domain.Snippet, title, content string) snippet.UpdateInput {
	return snippet.UpdateInput{
		SnippetID:       stored.ID(),
		LoadedUpdatedAt: stored.UpdatedAt(),
		Title:           title,
		Description:     stored.Description().String(),
		Content:         content,
	}
}
