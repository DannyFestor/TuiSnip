package tui_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/screencell"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	secondTitle       = "Prune everything"
	secondDescription = "Reclaim disk space"
	snippetPaneHint   = "w wrap"
)

func TestModel_mouse(t *testing.T) {
	t.Parallel()

	t.Run("asks the terminal for clicks and the wheel", func(t *testing.T) {
		t.Parallel()

		screen := start(t, mouseModel(t, testkit.NewManualClock(time.Time{}), true), wideWidth, wideHeight)

		assert.Equal(t, tea.MouseModeCellMotion, screen.model.View().MouseMode)
	})

	t.Run("a click selects the Snippet under it", func(t *testing.T) {
		t.Parallel()

		screen := start(t, mouseModel(t, testkit.NewManualClock(time.Time{}), true), wideWidth, wideHeight)

		screen.send(clickOnSecondTitle(t, screen))

		assert.Contains(t, screen.screen(), secondDescription)
	})

	t.Run("two quick clicks on one Snippet open it in the Snippet pane", func(t *testing.T) {
		t.Parallel()

		clock := testkit.NewManualClock(time.Time{})
		screen := start(t, mouseModel(t, clock, true), wideWidth, wideHeight)

		click := clickOnSecondTitle(t, screen)

		screen.send(click)
		clock.Advance(200 * time.Millisecond)
		screen.send(click)

		assert.Contains(t, screen.screen(), snippetPaneHint)
	})

	t.Run("two slow clicks on one Snippet stay in the Snippet list", func(t *testing.T) {
		t.Parallel()

		clock := testkit.NewManualClock(time.Time{})
		screen := start(t, mouseModel(t, clock, true), wideWidth, wideHeight)

		click := clickOnSecondTitle(t, screen)

		screen.send(click)
		clock.Advance(time.Second)
		screen.send(click)

		assert.NotContains(t, screen.screen(), snippetPaneHint)
	})

	t.Run("the wheel scrolls the Snippet list under the pointer", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, listLongerThanPane)
		screen := start(t, newModel(t, listerOf(t, snippets...), NewMockSnippetCopier(t)), wideWidth, wideHeight)
		at := screencell.Find(t, screen.screen(), "Snippet 2 ")

		screen.send(tea.MouseWheelMsg{X: at.X, Y: at.Y, Button: tea.MouseWheelDown, Mod: 0})

		assert.Equal(t, at, screencell.Find(t, screen.screen(), "Snippet 5 "))
	})
}

func TestModel_mouseOff(t *testing.T) {
	t.Parallel()

	t.Run("asks the terminal for no mouse reports", func(t *testing.T) {
		t.Parallel()

		screen := start(t, mouseModel(t, testkit.NewManualClock(time.Time{}), false), wideWidth, wideHeight)

		assert.Equal(t, tea.MouseModeNone, screen.model.View().MouseMode)
	})

	t.Run("ignores a click", func(t *testing.T) {
		t.Parallel()

		screen := start(t, mouseModel(t, testkit.NewManualClock(time.Time{}), false), wideWidth, wideHeight)

		screen.send(clickOnSecondTitle(t, screen))

		assert.NotContains(t, screen.screen(), secondDescription)
	})

	t.Run("ignores the wheel", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, listLongerThanPane)
		settings := testsettings.Default(t)
		settings.Mouse = false
		model := modelWithSettings(t, actions{
			lister:     listerOf(t, snippets...),
			treeLister: treeOf(t, emptyTree()),
			copier:     NewMockSnippetCopier(t),
			creator:    NewMockSnippetCreator(t),
			searcher:   NewMockSnippetSearcher(t),
		}, settings)
		screen := start(t, model, wideWidth, wideHeight)
		at := screencell.Find(t, screen.screen(), "Snippet 2 ")

		screen.send(tea.MouseWheelMsg{X: at.X, Y: at.Y, Button: tea.MouseWheelDown, Mod: 0})

		assert.Equal(t, at, screencell.Find(t, screen.screen(), "Snippet 2 "))
	})
}

func mouseModel(t *testing.T, clock tui.Clock, mouse bool) tui.Model {
	t.Helper()

	settings := testsettings.Default(t)
	settings.Mouse = mouse

	return modelWithSettings(t, actions{
		lister:     listerOf(t, sampleSnippets(t)...),
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
		clock:      clock,
	}, settings)
}

func clickOnSecondTitle(t *testing.T, screen *driver) tea.MouseClickMsg {
	t.Helper()

	at := screencell.Find(t, screen.screen(), secondTitle)

	return tea.MouseClickMsg{X: at.X, Y: at.Y, Button: tea.MouseLeft, Mod: 0}
}
