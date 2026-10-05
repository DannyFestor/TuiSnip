//go:build feature

package feature_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

const (
	keepTrailingNewline = "[copy]\ntrim_trailing_newline = false\n"
	nativeOnly          = "[copy]\nclipboard = \"native\"\n"
	osc52Only           = "[copy]\nclipboard = \"osc52\"\n"
	contentCapBytes     = 256 << 10
)

func TestCopyTrimsTrailingNewlineByDefault(t *testing.T) {
	t.Parallel()

	home, app := testapp.Start(t, testapp.RecordingTool)
	created := create(
		t,
		app,
		snippet.CreateInput{Title: "ls", Description: "", Language: plainText, Content: "ls -la\n"},
	)

	result := copySnippet(t, app, created.ID())

	assert.Equal(t, domain.CopyDeliveryPlaced, result.Delivery)
	assert.Equal(t, "ls -la", home.Copied(t))
}

func TestCopyPlacesContentByteExact(t *testing.T) {
	t.Parallel()

	content := "echo 'Grüße 🚀'\r\nprintf '\\t'\n\n"
	home := testapp.NewHome(t)
	home.WriteConfig(t, keepTrailingNewline)
	app := home.Start(t, testapp.RecordingTool)
	created := create(
		t,
		app,
		snippet.CreateInput{Title: "unicode", Description: "", Language: plainText, Content: content},
	)

	copySnippet(t, app, created.ID())

	assert.Equal(t, content, home.Copied(t))
}

func TestCopyAtContentCap(t *testing.T) {
	t.Parallel()

	content := strings.Repeat("x", contentCapBytes)
	home := testapp.NewHome(t)
	home.WriteConfig(t, keepTrailingNewline)
	app := home.Start(t, testapp.RecordingTool)
	created := create(
		t,
		app,
		snippet.CreateInput{Title: "large", Description: "", Language: plainText, Content: content},
	)

	copySnippet(t, app, created.ID())

	assert.Len(t, home.Copied(t), contentCapBytes)
}

func TestCopyOfMissingSnippetIsNotFound(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	removed := create(t, app, snippet.CreateInput{Title: "ls", Description: "", Language: plainText, Content: "ls"})
	_, otherApp := testapp.Start(t, testapp.RecordingTool)

	_, err := otherApp.Copy.Run(t.Context(), snippet.CopyInput{SnippetID: removed.ID()})

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestNativeCopyWithoutToolFails(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, nativeOnly)
	app := home.Start(t, testapp.NoTool)
	created := create(t, app, snippet.CreateInput{Title: "ls", Description: "", Language: plainText, Content: "ls"})

	_, err := app.Copy.Run(t.Context(), snippet.CopyInput{SnippetID: created.ID()})

	require.ErrorIs(t, err, domain.ErrNoClipboardTool)
}

func TestAutoCopySendsToTerminalWhenToolFails(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.FailingTool)
	created := create(t, app, snippet.CreateInput{Title: "ls", Description: "", Language: plainText, Content: "ls\n"})

	result := copySnippet(t, app, created.ID())

	assert.Equal(t, domain.CopyDeliverySentToTerminal, result.Delivery)
	assert.Equal(t, "ls", result.Content.String())
}

func TestOSC52CopySendsToTerminal(t *testing.T) {
	t.Parallel()

	home := testapp.NewHome(t)
	home.WriteConfig(t, osc52Only)
	app := home.Start(t, testapp.RecordingTool)
	created := create(t, app, snippet.CreateInput{Title: "ls", Description: "", Language: plainText, Content: "ls"})

	result := copySnippet(t, app, created.ID())

	assert.Equal(t, domain.CopyDeliverySentToTerminal, result.Delivery)
	assert.NoFileExists(t, home.ToolRecording())
}

func copySnippet(t *testing.T, app *bootstrap.App, id domain.SnippetID) snippet.CopyResult {
	t.Helper()

	result, err := app.Copy.Run(t.Context(), snippet.CopyInput{SnippetID: id})
	require.NoError(t, err)

	return result
}
