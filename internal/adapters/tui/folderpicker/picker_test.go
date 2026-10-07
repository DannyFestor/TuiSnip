package folderpicker_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/arrived"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/foldertree"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestPicker_View(t *testing.T) {
	t.Parallel()

	t.Run("lists the Root and every Folder by path in tree order", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, foldertree.New(t), domain.FolderID{}, domain.FolderID{})

		assert.Contains(t, screen.Screen(), pickerTitle)
		assert.Regexp(t, `(?s)Root\s.*\n.*docker\s.*\n.*go\s.*\n.*go / testing`, screen.Screen())
	})

	t.Run("hints move, pick and close", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, foldertree.New(t), domain.FolderID{}, domain.FolderID{})

		assert.Equal(t, "down move · enter pick · esc close", screen.Hints())
	})

	t.Run("draws the filter and the Folders in the new Styles", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		light := look.NewStyles(look.SchemeLight)
		screen := picking(t, sample, domain.FolderID{}, domain.FolderID{})
		screen.Press(keypress.Typed("go")...)

		screen.Send(look.Restyled{Styles: light})

		styledFromStart := pickingStyled(t, light, offerOf(sample.Tree, domain.FolderID{}, domain.FolderID{}))
		styledFromStart.Press(keypress.Typed("go")...)
		assert.Equal(t, styledFromStart.StyledScreen(), screen.StyledScreen())
	})
}

func TestPicker_Update(t *testing.T) {
	t.Parallel()

	t.Run("picks the highlighted Folder and closes", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, domain.FolderID{}, domain.FolderID{})

		screen.Press(down(), enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Docker.ID())}, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("picks the Root", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, sample.Tests.ID(), domain.FolderID{})

		screen.Press(keypress.Typed("root")...)
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(domain.FolderID{})}, screen.Outcomes())
	})

	t.Run("opens on the current Folder", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, sample.Go.ID(), domain.FolderID{})

		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Go.ID())}, screen.Outcomes())
	})

	t.Run("picks from the Folders whose path matches the filter", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, domain.FolderID{}, domain.FolderID{})

		screen.Press(keypress.Typed("test")...)
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Tests.ID())}, screen.Outcomes())
	})

	t.Run("refuses the moving Folder and every Folder below it", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)

		for _, refused := range []string{"go", "go / testing"} {
			screen := picking(t, sample, domain.FolderID{}, sample.Go.ID())

			screen.Press(keypress.Typed(refused)...)
			screen.Press(enter())

			assert.Empty(t, screen.Outcomes(), refused)
			assert.True(t, screen.IsOpen(), refused)
		}
	})

	t.Run("still offers the Folders outside the moving Folder's subtree", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, domain.FolderID{}, sample.Go.ID())

		screen.Press(down(), enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Docker.ID())}, screen.Outcomes())
	})

	t.Run("closes without picking", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, foldertree.New(t), domain.FolderID{}, domain.FolderID{})

		screen.Press(keypress.Special(tea.KeyEscape))

		assert.Empty(t, screen.Outcomes())
		assert.False(t, screen.IsOpen())
	})

	t.Run("matches the typed filter against a Tree that arrives while open", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := pickingIn(t, browse.Tree{}, domain.FolderID{}, domain.FolderID{})

		screen.Press(keypress.Typed("docker")...)
		screen.Send(arrived.Tree{Tree: sample.Tree})
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Docker.ID())}, screen.Outcomes())
	})

	t.Run("refuses the moving Folder's subtree in a Tree that arrives while open", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)

		for _, refused := range []string{"go", "go / testing"} {
			screen := pickingIn(t, browse.Tree{}, domain.FolderID{}, sample.Go.ID())

			screen.Send(arrived.Tree{Tree: sample.Tree})
			screen.Press(keypress.Typed(refused)...)
			screen.Press(enter())

			assert.Empty(t, screen.Outcomes(), refused)
			assert.True(t, screen.IsOpen(), refused)
		}
	})

	t.Run("lands on the current Folder once its Tree arrives", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := pickingIn(t, browse.Tree{}, sample.Go.ID(), domain.FolderID{})

		screen.Send(arrived.Tree{Tree: sample.Tree})
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Go.ID())}, screen.Outcomes())
	})

	t.Run("keeps the Folder the user moved to when the Tree arrives again", func(t *testing.T) {
		t.Parallel()

		sample := foldertree.New(t)
		screen := picking(t, sample, domain.FolderID{}, domain.FolderID{})

		screen.Press(down())
		screen.Send(arrived.Tree{Tree: sample.Tree})
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{pickedOutcome(sample.Docker.ID())}, screen.Outcomes())
	})

	t.Run("stays open when nothing matches the filter", func(t *testing.T) {
		t.Parallel()

		screen := picking(t, foldertree.New(t), domain.FolderID{}, domain.FolderID{})

		screen.Press(keypress.Typed("zzz")...)
		screen.Press(enter())

		assert.Contains(t, screen.Screen(), noMatches)
		assert.Empty(t, screen.Outcomes())
		assert.True(t, screen.IsOpen())
	})
}
