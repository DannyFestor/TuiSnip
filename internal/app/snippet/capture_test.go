package snippet_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestNewCapture(t *testing.T) {
	t.Parallel()

	_, err := snippet.NewCapture(nil)

	require.ErrorIs(t, err, domain.ErrMissingDependency)
	require.ErrorContains(t, err, "snippet.NewCapture: ")
	assert.ErrorContains(t, err, "clipboard")
}

func TestCapture_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns the clipboard's content as it is", func(t *testing.T) {
		t.Parallel()

		content, err := newCapture(t, readingClipboard(t, "\tdocker ps\n")).Run(t.Context(), snippet.CaptureInput{})

		require.NoError(t, err)
		assert.Equal(t, "\tdocker ps\n", content.String())
	})

	tests := []struct {
		name    string
		read    string
		readErr error
		wantErr error
	}{
		{name: "refuses an empty clipboard", read: "", wantErr: domain.ErrClipboardEmpty},
		{
			name:    "refuses content larger than a Fragment holds",
			read:    strings.Repeat("c", maxContentBytes+1),
			wantErr: value.ErrContentTooLong,
		},
		{name: "returns the read error", readErr: domain.ErrNoClipboardTool, wantErr: domain.ErrNoClipboardTool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clipboard := NewMockClipboardReader(t)
			clipboard.EXPECT().ReadClipboard(mock.Anything).Return(tt.read, tt.readErr)

			_, err := newCapture(t, clipboard).Run(t.Context(), snippet.CaptureInput{})

			require.ErrorIs(t, err, tt.wantErr)
			assert.ErrorContains(t, err, "snippet.Capture: ")
		})
	}
}

func newCapture(t *testing.T, clipboard snippet.ClipboardReader) *snippet.Capture {
	t.Helper()

	capture, err := snippet.NewCapture(clipboard)
	require.NoError(t, err)

	return capture
}

func readingClipboard(t *testing.T, content string) *MockClipboardReader {
	t.Helper()

	clipboard := NewMockClipboardReader(t)
	clipboard.EXPECT().ReadClipboard(mock.Anything).Return(content, nil)

	return clipboard
}
