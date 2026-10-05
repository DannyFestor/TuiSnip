package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	tabbedContent   = "if x {\n\treturn\n}\n"
	readOnlyNotice  = "Contains tabs: read-only here, edit with ctrl+e ($EDITOR)"
	unboundReadOnly = "Contains tabs: read-only here"
)

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

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune all", "Reclaim", "echo hi!")}
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

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune all", "Reclaim", tabbedContent)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("closes and reports the saved Snippet", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Press(keypress.Letter('x'), save())
		screen.Send(editoverlay.SaveFinished{Snippet: stored, Err: nil})

		reported := outcome.SnippetSaved{ID: stored.ID(), FolderID: stored.FolderID()}
		assert.Contains(t, screen.Outcomes(), outcome.Outcome(reported))
		assert.False(t, screen.IsOpen())
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

func updateInput(stored domain.Snippet, title, description, content string) snippet.UpdateInput {
	return snippet.UpdateInput{
		SnippetID:       stored.ID(),
		LoadedUpdatedAt: stored.UpdatedAt(),
		Title:           title,
		Description:     description,
		Content:         content,
	}
}
