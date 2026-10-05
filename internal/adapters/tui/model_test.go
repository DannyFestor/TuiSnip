package tui_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const snippetPaneTitle = "4 Snippet"

var errDatabaseLocked = errors.New("database is locked")

func TestNew(t *testing.T) {
	t.Parallel()

	constructors := []struct {
		name  string
		build modelConstructor
	}{
		{name: "New", build: tui.New},
		{name: "NewQuittingAfterCopy", build: tui.NewQuittingAfterCopy},
	}

	for _, tt := range constructors {
		t.Run(tt.name+" requires every dependency", func(t *testing.T) {
			t.Parallel()

			settings := testsettings.Default(t)
			settings.Location = nil

			_, err := tt.build(t.Context(), tui.Deps{
				Lister:     nil,
				TreeLister: nil,
				Copier:     nil,
				Creator:    nil,
				Searcher:   nil,
				Settings:   settings,
				Logger:     nil,
			})

			require.ErrorIs(t, err, domain.ErrMissingDependency)

			for _, name := range []string{
				"lister", "treeLister", "tagLister", "tagSnippetsLister", "copier", "creator", "updater", "searcher",
				"sortOrderSaver", "logger", "location",
			} {
				assert.ErrorContains(t, err, name)
			}
		})
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
		lister.EXPECT().Run(mock.Anything, atRootByTitle()).Return(nil, errDatabaseLocked)

		screen := start(t, newModel(t, lister, NewMockSnippetCopier(t)), wideWidth, wideHeight)

		assert.Contains(t, screen.screen(), "Something went wrong; see the log")
	})
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

			screen := copyFirstSnippet(t, tui.New, tt.result, tt.err)

			assert.Contains(t, screen.screen(), tt.status)
		})
	}

	t.Run("sends the content to the terminal over OSC 52", func(t *testing.T) {
		t.Parallel()

		snippets := sampleSnippets(t)
		copier := NewMockSnippetCopier(t)
		copier.EXPECT().Run(mock.Anything, mock.Anything).Return(copied(t, domain.CopyDeliverySentToTerminal), nil)
		screen := start(t, newModel(t, listerOf(t, snippets...), copier), wideWidth, wideHeight)

		screen.press(keypress.Letter('4'), keypress.Letter('y'))

		assert.Contains(t, screen.emitted, tea.SetClipboard("echo copied")())
	})

	t.Run("stays quiet when cancelled", func(t *testing.T) {
		t.Parallel()

		screen := copyFirstSnippet(t, tui.New, snippet.CopyResult{}, context.Canceled)

		assert.NotContains(t, screen.screen(), "Something went wrong")
	})
}

func TestModel_quitAfterCopy(t *testing.T) {
	t.Parallel()

	t.Run("quits after a Copy placed on the clipboard", func(t *testing.T) {
		t.Parallel()

		screen := copyFirstSnippet(t, tui.NewQuittingAfterCopy, copied(t, domain.CopyDeliveryPlaced), nil)

		assert.Contains(t, screen.emitted, tea.QuitMsg{})
	})

	t.Run("sends the content to the terminal before quitting", func(t *testing.T) {
		t.Parallel()

		screen := copyFirstSnippet(t, tui.NewQuittingAfterCopy, copied(t, domain.CopyDeliverySentToTerminal), nil)

		clipboard := slices.Index(screen.emitted, tea.SetClipboard("echo copied")())
		quit := slices.Index(screen.emitted, tea.Msg(tea.QuitMsg{}))

		require.NotEqual(t, -1, clipboard)
		require.NotEqual(t, -1, quit)
		assert.Less(t, clipboard, quit)
	})

	failures := []struct {
		name string
		err  error
	}{
		{name: "stays open when no clipboard tool exists", err: domain.ErrNoClipboardTool},
		{name: "stays open after a failed Copy", err: errDatabaseLocked},
		{name: "stays open after a cancelled Copy", err: context.Canceled},
	}

	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			screen := copyFirstSnippet(t, tui.NewQuittingAfterCopy, snippet.CopyResult{}, tt.err)

			assert.NotContains(t, screen.emitted, tea.QuitMsg{})
		})
	}

	t.Run("stays open after a successful Copy when the setting is off", func(t *testing.T) {
		t.Parallel()

		screen := copyFirstSnippet(t, tui.New, copied(t, domain.CopyDeliveryPlaced), nil)

		assert.NotContains(t, screen.emitted, tea.QuitMsg{})
		assert.Contains(t, screen.screen(), "Copied")
	})
}

func TestModel_quit(t *testing.T) {
	t.Parallel()

	for name, pressed := range map[string]tea.KeyPressMsg{"q": keypress.Letter('q'), "ctrl+c": keypress.Ctrl('c')} {
		t.Run(name+" quits", func(t *testing.T) {
			t.Parallel()

			screen := start(t, newModel(t, listerOf(t), NewMockSnippetCopier(t)), wideWidth, wideHeight)

			screen.press(pressed)

			assert.Contains(t, screen.emitted, tea.QuitMsg{})
		})
	}
}

func copyFirstSnippet(t *testing.T, build modelConstructor, result snippet.CopyResult, err error) *driver {
	t.Helper()

	snippets := sampleSnippets(t)
	copier := NewMockSnippetCopier(t)
	copier.EXPECT().Run(mock.Anything, snippet.CopyInput{SnippetID: snippets[0].ID()}).Return(result, err)

	model := modelBuiltBy(t, build, actions{
		lister:     listerOf(t, snippets...),
		treeLister: treeOf(t, emptyTree()),
		copier:     copier,
		creator:    NewMockSnippetCreator(t),
		searcher:   NewMockSnippetSearcher(t),
	}, testsettings.Default(t))
	screen := start(t, model, wideWidth, wideHeight)

	screen.press(keypress.Letter('3'), keypress.Letter('y'))

	return screen
}
