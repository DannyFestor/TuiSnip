package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	externallyEditedContent = "echo edited\n"
	overlongReadOnlyNotice  = "Longer than 10,000 lines: read-only here, edit with ctrl+e ($EDITOR)"
)

func TestSession_Update_externalEditor(t *testing.T) {
	t.Parallel()

	t.Run("asks for the external editor from a field with the Fragment's content and Language", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, "echo hi"))
		screen.Press(externalEditor())

		want := externalEditAsked("echo hi", language(t, "Go"))
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("asks for the external editor from Content with what was typed", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("ls -la")...)
		screen.Press(externalEditor())

		want := externalEditAsked("ls -la", value.PlainText())
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("asks for the external editor on read-only content", func(t *testing.T) {
		t.Parallel()

		screen := editingStored(t, storedSnippet(t, tabbedContent))
		screen.Press(toContent()...)
		screen.Press(externalEditor())

		want := externalEditAsked(tabbedContent, language(t, "Go"))
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})
}

func TestSession_Received_contentEdited(t *testing.T) {
	t.Parallel()

	t.Run("puts the edited content in as an unsaved change", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Offer(contentEdited(externallyEditedContent))

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "echo edited")

		screen.Press(save())

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", externallyEditedContent)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("lands edited content with tabs read-only and saves it byte for byte", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Offer(contentEdited(tabbedContent))
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		assert.Contains(t, screen.Screen(), readOnlyNotice)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", tabbedContent)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("makes content editable again once the edit removed its tabs", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, tabbedContent)
		screen := editingStored(t, stored)
		screen.Offer(contentEdited(externallyEditedContent))
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		assert.NotContains(t, screen.Screen(), readOnlyNotice)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", externallyEditedContent+"!")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("lands edited content over 10,000 lines read-only and saves it byte for byte", func(t *testing.T) {
		t.Parallel()

		overlong := linesOf(10_001)
		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Press(enterContent()...)
		screen.Offer(contentEdited(overlong))
		screen.Press(keypress.Typed("lost")...)
		screen.Press(save())

		assert.Contains(t, screen.Screen(), overlongReadOnlyNotice)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", overlong)}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("keeps edited content of exactly 10,000 lines editable", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Offer(contentEdited(linesOf(10_000)))
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		assert.NotContains(t, screen.Screen(), overlongReadOnlyNotice)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", linesOf(10_000)+"!")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})

	t.Run("makes content editable again once the edit brought it to 10,000 lines", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi")
		screen := editingStored(t, stored)
		screen.Offer(contentEdited(linesOf(10_001)))
		screen.Offer(contentEdited(externallyEditedContent))
		screen.Press(enterContent()...)
		screen.Press(keypress.Typed("!")...)
		screen.Press(save())

		assert.NotContains(t, screen.Screen(), overlongReadOnlyNotice)

		want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", externallyEditedContent+"!")}
		assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
	})
}

func TestEditedExternally(t *testing.T) {
	t.Parallel()

	stored := storedSnippet(t, "echo hi")
	opened, _ := editoverlay.EditedExternally(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		nil,
		editoverlay.ExternallyEdited{
			Browsed: editoverlay.BrowsedSnippet{Snippet: stored, Selection: browsed()},
			Content: externallyEditedContent,
		},
	)
	screen := overlaytest.Open(t, screenSize(), opened)

	assert.Contains(t, screen.Screen(), unsavedTitle)
	assert.Contains(t, screen.Screen(), "echo edited")

	screen.Press(save())

	want := outcome.UpdateRequested{Input: updateInput(stored, "Prune", externallyEditedContent)}
	assert.Equal(t, []outcome.Outcome{want}, screen.Outcomes())
}

func externalEditor() tea.KeyPressMsg {
	return keypress.Ctrl('e')
}

func externalEditAsked(content string, in value.Language) outcome.ExternalEditAsked {
	return outcome.ExternalEditAsked{
		Content:   content,
		Language:  in,
		Snippet:   domain.Snippet{},
		Selection: browseselection.Selection{},
	}
}

func contentEdited(content string) outcome.ContentEdited {
	return outcome.ContentEdited{Asked: externalEditAsked("echo hi", value.PlainText()), Content: content}
}
