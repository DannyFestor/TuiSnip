//go:build e2e

package e2e_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	rootTitleInRowHeaderAndStatus = 3
	rootTitleInTwoRowsAndHeader   = 3
	rootTitleMovedToTrash         = `Moved Snippet "Prune at the Root" to the Trash`
	emptySnippetListHint          = "No Snippets here."
)

func TestDuplicateAndDeleteSnippets(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: rootTitle})
	screen := open(t, app)
	screen.waitForFrame(rootTitle)

	screen.press(keypress.Letter('3'), keypress.Letter('c'))
	screen.waitForCount(rootTitle, rootTitleInTwoRowsAndHeader)

	screen.press(keypress.Letter('d'))
	screen.waitForFrame(rootTitleMovedToTrash)
	screen.waitForCount(rootTitle, rootTitleInRowHeaderAndStatus)

	screen.press(keypress.Letter('d'))
	screen.waitForFrame(emptySnippetListHint)
}

func (s *session) waitForCount(text string, count int) {
	s.t.Helper()

	require.Eventually(s.t, func() bool {
		return strings.Count(s.frame.get(), text) == count
	}, waitTimeout, pollInterval, "%q never appeared %d times", text, count)
}
