package mainscreen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/mainscreen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestScreen_duplicate(t *testing.T) {
	t.Parallel()

	t.Run("c in the Snippet list asks to duplicate the selected Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := showing(t, wide(), snippets...)

		screen.Press(keypress.Typed("3jc")...)

		assert.Equal(t, []outcome.Outcome{outcome.SnippetDuplicateRequested{
			Input:     snippet.DuplicateInput{SnippetID: snippets[1].ID()},
			Title:     snippets[1].Title().String(),
			Selection: browseselection.Selection{},
		}}, screen.Outcomes())
	})

	t.Run("c asks for the selected Folder", func(t *testing.T) {
		t.Parallel()

		screen, sample := browsing(t)
		screen.Press(keypress.Typed("jj")...)
		screen.Send(mainscreen.SnippetsLoaded{
			Selection: browseselection.InFolder(sample.Go.ID()),
			Snippets:  []domain.Snippet{filedIn(t, sample.Go.ID())},
			Order:     domain.SortOrderTitle,
		})

		screen.Press(keypress.Typed("3c")...)

		assert.Contains(t, screen.Outcomes(), outcome.SnippetDuplicateRequested{
			Input:     snippet.DuplicateInput{SnippetID: filedIn(t, sample.Go.ID()).ID()},
			Title:     filedIn(t, sample.Go.ID()).Title().String(),
			Selection: browseselection.InFolder(sample.Go.ID()),
		})
	})

	t.Run("c does nothing outside the Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("c4c")...)

		assert.Empty(t, screen.Outcomes())
	})

	t.Run("c does nothing in an empty Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Typed("3c")...)

		assert.Empty(t, screen.Outcomes())
	})
}

func TestScreen_snippetDelete(t *testing.T) {
	t.Parallel()

	t.Run("d in the Snippet list moves it to the Trash without asking", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide(), sampleSnippets(t)...)

		screen.Press(keypress.Typed("3jd")...)

		assert.NotContains(t, screen.Screen(), "[y/N]")
		assert.Len(t, screen.Outcomes(), 1)
	})

	selectingAfter := []struct {
		name      string
		typed     string
		deleted   int
		selecting int
	}{
		{name: "asks to delete the Snippet, selecting the one below it", typed: "3d", deleted: 0, selecting: 1},
		{name: "on the last Snippet selects the one above it", typed: "3jd", deleted: 1, selecting: 0},
	}

	for _, tt := range selectingAfter {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippets := sampleSnippets(t)
			screen := showing(t, wide(), snippets...)

			screen.Press(keypress.Typed(tt.typed)...)

			assert.Equal(t, []outcome.Outcome{outcome.SnippetDeleteRequested{
				Input:     snippet.DeleteInput{SnippetID: snippets[tt.deleted].ID()},
				Title:     snippets[tt.deleted].Title().String(),
				Selection: browseselection.Selection{},
				Selecting: snippets[tt.selecting].ID(),
			}}, screen.Outcomes())
		})
	}

	t.Run("on the only Snippet selects nothing", func(t *testing.T) {
		t.Parallel()

		only := sampleSnippets(t)[0]
		screen := showing(t, wide(), only)

		screen.Press(keypress.Typed("3d")...)

		assert.Equal(t, []outcome.Outcome{outcome.SnippetDeleteRequested{
			Input:     snippet.DeleteInput{SnippetID: only.ID()},
			Title:     only.Title().String(),
			Selection: browseselection.Selection{},
			Selecting: domain.SnippetID{},
		}}, screen.Outcomes())
	})

	t.Run("d does nothing in an empty Snippet list", func(t *testing.T) {
		t.Parallel()

		screen := showing(t, wide())

		screen.Press(keypress.Typed("3d")...)

		assert.Empty(t, screen.Outcomes())
	})
}
