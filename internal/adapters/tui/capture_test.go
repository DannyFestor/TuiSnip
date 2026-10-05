package tui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const (
	clipboardText         = "docker ps --all\n"
	clipboardEmptyText    = "Clipboard is empty"
	noCaptureToolText     = "No clipboard tool found (pbpaste, wl-paste, xclip, xsel)"
	clipboardTooLargeText = "Clipboard is larger than 256 KiB"
)

func TestModel_capture(t *testing.T) {
	t.Parallel()

	t.Run("opens the clipboard in the edit overlay and saves it", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 1)
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: "List", Content: clipboardText}).
			Return(snippets[0], nil)
		with := capturingActions(t, capturerReading(t, clipboardText), listerReturning(t, nil, snippets))
		with.creator = creator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('p'))

		assert.Contains(t, screen.screen(), editOverlayTitle+" •")
		assert.Contains(t, screen.screen(), "docker ps --all")

		screen.press(keypress.Typed("List")...)
		screen.press(keypress.Ctrl('s'))

		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})

	tests := []struct {
		name     string
		err      error
		wantText string
	}{
		{name: "says the clipboard is empty", err: domain.ErrClipboardEmpty, wantText: clipboardEmptyText},
		{name: "names the tools it looked for", err: domain.ErrNoClipboardTool, wantText: noCaptureToolText},
		{name: "says the clipboard is over 256 KiB", err: value.ErrContentTooLong, wantText: clipboardTooLargeText},
		{name: "shows a generic failure for anything else", err: errDatabaseLocked, wantText: genericFailure},
	}
	for _, tt := range tests {
		t.Run(tt.name+" and opens nothing", func(t *testing.T) {
			t.Parallel()

			capturer := NewMockSnippetCapturer(t)
			capturer.EXPECT().
				Run(mock.Anything, snippet.CaptureInput{}).
				Return(value.Content{}, fmt.Errorf("snippet.Capture: %w", tt.err))
			screen := start(t, modelWith(t, capturingActions(t, capturer, listerOf(t))), wideWidth, wideHeight)

			screen.press(keypress.Letter('p'))

			assert.Contains(t, screen.screen(), tt.wantText)
			assert.NotContains(t, screen.screen(), editOverlayTitle)
		})
	}

	t.Run("refuses a clipboard over 10,000 lines and opens nothing", func(t *testing.T) {
		t.Parallel()

		overlong := strings.Repeat("line\n", 10_000)
		screen := start(
			t,
			modelWith(t, capturingActions(t, capturerReading(t, overlong), listerOf(t))),
			wideWidth,
			wideHeight,
		)

		screen.press(keypress.Letter('p'))

		assert.Contains(t, screen.screen(), "Clipboard is longer than 10,000 lines; use ctrl+e to edit in $EDITOR")
		assert.NotContains(t, screen.screen(), editOverlayTitle)
	})

	t.Run("keeps the Tag as the Browse selection after saving a new Snippet for it", func(t *testing.T) {
		t.Parallel()

		tags := sampleTagCounts(t)
		created := filedIn(t, domain.FolderID{})
		creator := NewMockSnippetCreator(t)
		creator.EXPECT().
			Run(mock.Anything, snippet.CreateInput{Title: filedTitle, Tags: []domain.Tag{tags[0].Tag}}).
			Return(created, nil)

		tagged := NewMockTagSnippetsLister(t)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle)
		listingWithTag(tagged, tags[0].Tag.ID(), domain.SortOrderTitle, created)

		with := taggedActions(t, listerOf(t), tagsOf(t, tags...), tagged)
		with.creator = creator
		screen := start(t, modelWith(t, with), wideWidth, wideHeight)

		screen.press(keypress.Letter('2'), keypress.Special(tea.KeyEnter))
		screen.press(keypress.Letter('n'))
		screen.press(keypress.Typed(filedTitle)...)
		screen.press(keypress.Ctrl('s'))

		assert.Contains(t, screen.screen(), "3 # docker · by title")
		assert.Contains(t, screen.screen(), filedTitle)
	})
}

func capturingActions(t *testing.T, capturer tui.SnippetCapturer, lister *MockFolderSnippetsLister) actions {
	t.Helper()

	return actions{
		lister:     lister,
		treeLister: treeOf(t, emptyTree()),
		copier:     NewMockSnippetCopier(t),
		creator:    NewMockSnippetCreator(t),
		capturer:   capturer,
		searcher:   NewMockSnippetSearcher(t),
	}
}

func capturerReading(t *testing.T, text string) *MockSnippetCapturer {
	t.Helper()

	content, err := value.NewContent(text)
	require.NoError(t, err)

	capturer := NewMockSnippetCapturer(t)
	capturer.EXPECT().Run(mock.Anything, snippet.CaptureInput{}).Return(content, nil)

	return capturer
}
