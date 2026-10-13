package mainscreen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestScreen_externalEditor(t *testing.T) {
	t.Parallel()

	t.Run("E in the Snippet list asks to edit the selected Snippet externally", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("3E")...)

		assert.Equal(t, []outcome.Outcome{externalEditAskedFor(snippets[0])}, screen.Outcomes())
	})

	t.Run("E in the Snippet pane asks to edit the shown Snippet externally", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("3j4E")...)

		assert.Equal(t, []outcome.Outcome{externalEditAskedFor(snippets[1])}, screen.Outcomes())
	})

	t.Run("E asks to edit the Shown Fragment externally", func(t *testing.T) {
		t.Parallel()

		several := severalFragmentsSnippet(t)
		screen := showing(t, wide(), several)

		screen.Press(keypress.Typed("4]E")...)

		assert.Equal(t, []outcome.Outcome{externalEditAskedForFragment(several, 1)}, screen.Outcomes())
	})

	t.Run("E asks for nothing outside the Snippet list and the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Letter('E'))

		assert.Empty(t, screen.Outcomes())
	})

	t.Run("E asks for nothing in an empty Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Typed("3E4E")...)

		assert.Empty(t, screen.Outcomes())
	})

	t.Run("opens the edit overlay with the edited content as an unsaved change", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Offer(outcome.ContentEdited{Asked: externalEditAskedFor(snippets[1]), Content: "docker image prune\n"})

		assert.Contains(t, screen.Screen(), "Editing •")
		assert.Contains(t, screen.Screen(), "› Title          "+secondTitle)
		assert.Contains(t, screen.Screen(), "docker image prune")
		assert.Equal(t, editorHint, screen.Hints())
	})

	t.Run("opens the edit overlay on the externally edited Fragment", func(t *testing.T) {
		t.Parallel()

		several := severalFragmentsSnippet(t)
		screen := showing(t, wide(), several)

		screen.Offer(
			outcome.ContentEdited{Asked: externalEditAskedForFragment(several, 1), Content: "on: pull_request\n"},
		)

		assert.Contains(t, screen.Screen(), "Editing •")
		assert.Contains(t, screen.Screen(), "YAML")
		assert.Contains(t, screen.Screen(), "on: pull_request")
	})
}

func externalEditAskedFor(stored domain.Snippet) outcome.ExternalEditAsked {
	return externalEditAskedForFragment(stored, 0)
}

func externalEditAskedForFragment(stored domain.Snippet, index int) outcome.ExternalEditAsked {
	fragment := stored.Fragments()[index]

	return outcome.ExternalEditAsked{
		Content:   fragment.Content().String(),
		Language:  fragment.Language(),
		Snippet:   stored,
		Fragment:  fragment,
		Selection: browseselection.InFolder(domain.FolderID{}),
	}
}
