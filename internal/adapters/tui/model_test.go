package tui_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	folderPaneTitle  = "1 Folders"
	tagPaneTitle     = "2 Tags"
	snippetListTitle = "3 Root · by title"
	snippetPaneTitle = "4 Snippet"
)

var errDatabaseLocked = errors.New("database is locked")

func TestNew(t *testing.T) {
	t.Parallel()

	settings := testsettings.Default(t)
	settings.Location = nil

	_, err := tui.New(t.Context(), tui.Deps{
		Lister:   nil,
		Copier:   nil,
		Creator:  nil,
		Searcher: nil,
		Settings: settings,
		Logger:   nil,
	})

	require.ErrorIs(t, err, domain.ErrMissingDependency)

	for _, name := range []string{"lister", "copier", "creator", "searcher", "logger", "location"} {
		assert.ErrorContains(t, err, name)
	}
}

func TestModel_start(t *testing.T) {
	t.Parallel()

	t.Run("lists the Snippets at the Root and shows the first", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		screen := start(t, newModel(t, listerOf(t, snippets...), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "Prune everything")
		assert.Contains(t, screen.screen(), "Stop accepting, drain, exit")
		assert.Contains(t, screen.screen(), "created 2026-09-06 · updated 2026-09-27")
	})

	t.Run("shows the empty hint when the Root has no Snippets", func(t *testing.T) {
		t.Parallel()

		screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "No Snippets here.")
		assert.Contains(t, screen.screen(), "n  new Snippet")
	})

	t.Run("reports a listing failure", func(t *testing.T) {
		t.Parallel()

		lister := NewMockFolderSnippetsLister(t)
		lister.EXPECT().Run(mock.Anything, browse.SnippetsInFolderInput{}).Return(nil, errDatabaseLocked)

		screen := start(t, newModel(t, lister, NewMockSnippetCopier(t)), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
}

func TestModel_navigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		keys  []tea.KeyPressMsg
		title string
	}{
		{name: "starts on Folders", keys: nil, title: folderPaneTitle},
		{name: "tab moves to Tags", keys: []tea.KeyPressMsg{special(tea.KeyTab)}, title: tagPaneTitle},
		{
			name:  "tab wraps from the Snippet pane to Folders",
			keys:  []tea.KeyPressMsg{letter('4'), special(tea.KeyTab)},
			title: folderPaneTitle,
		},
		{
			name:  "shift+tab wraps from Folders to the Snippet pane",
			keys:  []tea.KeyPressMsg{{Code: tea.KeyTab, Mod: tea.ModShift}},
			title: snippetPaneTitle,
		},
		{name: "l from Tags goes to the Snippet list", keys: keys('2', 'l'), title: snippetListTitle},
		{name: "h from the Snippet list returns to Folders", keys: keys('2', 'l', 'h'), title: folderPaneTitle},
		{name: "l stops at the Snippet pane", keys: keys('4', 'l'), title: snippetPaneTitle},
		{name: "h stops at the left column", keys: keys('2', 'h'), title: tagPaneTitle},
		{
			name:  "enter drills from Folders to the Snippet pane",
			keys:  []tea.KeyPressMsg{special(tea.KeyEnter), special(tea.KeyEnter)},
			title: snippetPaneTitle,
		},
		{
			name:  "enter does nothing in Tags",
			keys:  []tea.KeyPressMsg{letter('2'), special(tea.KeyEnter)},
			title: tagPaneTitle,
		},
		{
			name:  "esc backs out from the Snippet pane to Folders",
			keys:  []tea.KeyPressMsg{letter('4'), special(tea.KeyEscape), special(tea.KeyEscape)},
			title: folderPaneTitle,
		},
		{name: "3 focuses the Snippet list", keys: keys('3'), title: snippetListTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), narrowWidth, narrowHeight)

			screen.press(tt.keys...)

			assert.Contains(t, screen.screen(), tt.title)
			assert.Contains(t, screen.screen(), "Terminal too small for all four Panes (80×24)")
		})
	}
}

func TestModel_cursor(t *testing.T) {
	t.Parallel()

	t.Run("moving down in the Snippet list shows the next Snippet", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)),
			wideWidth,
			wideHeight,
		)

		screen.press(letter('3'), letter('j'))

		assert.Contains(t, screen.screen(), "Reclaim disk space")
		assert.NotContains(t, screen.screen(), "Stop accepting, drain, exit")
	})

	t.Run("moving down in Folders keeps the Snippet", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)),
			wideWidth,
			wideHeight,
		)

		screen.press(letter('j'))

		assert.Contains(t, screen.screen(), "Stop accepting, drain, exit")
	})
}

func TestModel_minimumSize(t *testing.T) {
	t.Parallel()

	screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), minimumWidth, minimumHeight)

	for _, title := range []string{folderPaneTitle, tagPaneTitle, snippetListTitle, snippetPaneTitle} {
		assert.Contains(t, screen.screen(), title)
	}
}

func TestModel_copy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result snippet.CopyResult
		err    error
		status string
	}{
		{name: "says Copied when placed", result: copied(t, domain.CopyDeliveryPlaced), status: "Copied"},
		{
			name:   "says Sent to terminal over OSC 52",
			result: copied(t, domain.CopyDeliverySentToTerminal),
			status: "Sent to terminal",
		},
		{
			name:   "names the clipboard tools when none exists",
			err:    domain.ErrNoClipboardTool,
			status: "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)",
		},
		{
			name:   "shows a generic failure otherwise",
			err:    errDatabaseLocked,
			status: "Something went wrong; see the log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippets := sampleSnippets(t)
			copier := NewMockSnippetCopier(t)
			copier.EXPECT().Run(mock.Anything, snippet.CopyInput{SnippetID: snippets[0].ID()}).Return(tt.result, tt.err)
			screen := start(t, newModel(t, listerOf(t, snippets...), copier), wideWidth, wideHeight)

			screen.press(letter('3'), letter('y'))

			assert.Contains(t, screen.screen(), tt.status)
		})
	}

	t.Run("sends the content to the terminal over OSC 52", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		copier := NewMockSnippetCopier(t)
		copier.EXPECT().Run(mock.Anything, mock.Anything).Return(copied(t, domain.CopyDeliverySentToTerminal), nil)
		screen := start(t, newModel(t, listerOf(t, snippets...), copier), wideWidth, wideHeight)

		screen.press(letter('4'), letter('y'))

		assert.Contains(t, screen.emitted, tea.SetClipboard("echo copied")())
	})

	t.Run("stays quiet when cancelled", func(t *testing.T) {
		t.Parallel()

		copier := NewMockSnippetCopier(t)
		copier.EXPECT().Run(mock.Anything, mock.Anything).Return(snippet.CopyResult{}, context.Canceled)
		screen := start(t, newModel(t, listerOf(t, sampleSnippets(t)...), copier), wideWidth, wideHeight)

		screen.press(letter('3'), letter('y'))

		assert.NotContains(t, screen.screen(), "Something went wrong")
	})

	t.Run("does nothing outside the Snippet list and the Snippet pane", func(t *testing.T) {
		t.Parallel()

		screen := start(
			t,
			newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)),
			wideWidth,
			wideHeight,
		)

		screen.press(letter('y'), letter('2'), letter('y'))

		assert.NotContains(t, screen.screen(), "Copied")
	})
}

func TestModel_quit(t *testing.T) {
	t.Parallel()

	for name, pressed := range map[string]tea.KeyPressMsg{"q": letter('q'), "ctrl+c": ctrl('c')} {
		t.Run(name+" quits", func(t *testing.T) {
			t.Parallel()

			screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

			screen.press(pressed)

			assert.Contains(t, screen.emitted, tea.QuitMsg{})
		})
	}
}

func copied(t *testing.T, delivery domain.CopyDelivery) snippet.CopyResult {
	t.Helper()

	content, err := value.NewContent("echo copied")
	require.NoError(t, err)

	return snippet.CopyResult{Delivery: delivery, Content: content}
}

func keys(runes ...rune) []tea.KeyPressMsg {
	pressed := make([]tea.KeyPressMsg, 0, len(runes))
	for _, r := range runes {
		pressed = append(pressed, letter(r))
	}

	return pressed
}
