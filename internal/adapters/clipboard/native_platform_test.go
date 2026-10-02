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

	native := clipboard.NewNative(clipboard.Options{
		Environ:  os.Environ,
		LookPath: exec.LookPath,
		GOOS:     runtime.GOOS,
		Logger:   slog.New(slog.DiscardHandler),
	})

	delivery, err := native.Copy(t.Context(), unicodeText)

	require.NoError(t, err)
	assert.Equal(t, domain.CopyDeliveryPlaced, delivery)
	assert.Equal(t, unicodeText, pasteboard(t))
}

func keepClipboard(t *testing.T) {
	t.Helper()

	previous := pasteboard(t)
	ctx := context.WithoutCancel(t.Context())

	t.Cleanup(func() {
		restore := exec.CommandContext(ctx, "pbcopy")
		restore.Stdin = strings.NewReader(previous)
		require.NoError(t, restore.Run())
	})
}

func pasteboard(t *testing.T) string {
	t.Helper()

	out, err := exec.CommandContext(t.Context(), "pbpaste").Output()
	require.NoError(t, err)

	return string(out)
}
