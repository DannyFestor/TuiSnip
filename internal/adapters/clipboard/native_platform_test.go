//go:build darwin && platform

package clipboard_test

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

//nolint:paralleltest // the system clipboard is one shared resource
func TestNative_CopyThroughRealPbcopy(t *testing.T) {
	keepClipboard(t)

	delivery, err := realNative().Copy(t.Context(), unicodeText)

	require.NoError(t, err)
	assert.Equal(t, domain.CopyDeliveryPlaced, delivery)
	assert.Equal(t, unicodeText, pasteboard(t))
}

//nolint:paralleltest // the system clipboard is one shared resource
func TestNative_ReadThroughRealPbpaste(t *testing.T) {
	keepClipboard(t)
	placeOnPasteboard(t.Context(), t, unicodeText)

	read, err := realNative().ReadClipboard(t.Context())

	require.NoError(t, err)
	assert.Equal(t, unicodeText, read)
}

func realNative() *clipboard.Native {
	return clipboard.NewNative(clipboard.Options{
		Environ:  os.Environ,
		LookPath: exec.LookPath,
		GOOS:     runtime.GOOS,
		Logger:   slog.New(slog.DiscardHandler),
	})
}

func keepClipboard(t *testing.T) {
	t.Helper()

	previous := pasteboard(t)
	ctx := context.WithoutCancel(t.Context())

	t.Cleanup(func() { placeOnPasteboard(ctx, t, previous) })
}

func placeOnPasteboard(ctx context.Context, t *testing.T, text string) {
	t.Helper()

	place := exec.CommandContext(ctx, "pbcopy")
	place.Stdin = strings.NewReader(text)

	place.Env = append(os.Environ(), "LC_CTYPE=UTF-8")
	require.NoError(t, place.Run())
}

func pasteboard(t *testing.T) string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), "pbpaste").Output()
	require.NoError(t, err)

	return string(out)
}
