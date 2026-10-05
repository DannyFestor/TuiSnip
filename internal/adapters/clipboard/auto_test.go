package clipboard_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	fallbackLogMessage = "clipboard tool failed, sending OSC 52"
	concurrentCopies   = 8
)

func TestAuto_Copy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		env          []string
		script       string
		wantDelivery domain.CopyDelivery
		wantReceived string
		wantWarn     bool
	}{
		{
			name: "places the text with the tool on a local session", script: recordingScript,
			wantDelivery: domain.CopyDeliveryPlaced, wantReceived: unicodeText,
		},
		{
			name: "sends OSC 52 over SSH without running the tool", env: []string{"SSH_TTY=/dev/ttys001"},
			script: recordingScript, wantDelivery: domain.CopyDeliverySentToTerminal,
		},
		{
			name: "detects SSH from SSH_CONNECTION alone", env: []string{"SSH_CONNECTION=10.0.0.1 52000 10.0.0.2 22"},
			script: recordingScript, wantDelivery: domain.CopyDeliverySentToTerminal,
		},
		{
			name:         "sends OSC 52 and warns when no tool is installed",
			wantDelivery: domain.CopyDeliverySentToTerminal, wantWarn: true,
		},
		{
			name: "sends OSC 52 and warns when the tool fails", script: failingScript,
			wantDelivery: domain.CopyDeliverySentToTerminal, wantWarn: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := newFakeTools(t)
			if tt.script != "" {
				tools.install(t, "pbcopy", tt.script)
			}

			options, logs := withLogBuffer(tools.options(darwinGOOS, tt.env...))

			delivery, err := clipboard.NewAuto(options).Copy(t.Context(), unicodeText)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDelivery, delivery)
			assert.Equal(t, tt.wantReceived, tools.received(t, "pbcopy"))
			assert.Equal(t, tt.wantWarn, bytes.Contains(logs.Bytes(), []byte(fallbackLogMessage)))
		})
	}
}

func TestAuto_Copy_ToolHangs(t *testing.T) {
	t.Parallel()

	tools := newFakeTools(t)
	tools.install(t, "pbcopy", hangingScript)
	options, logs := withLogBuffer(tools.options(darwinGOOS))
	ctx, cancel := context.WithTimeout(t.Context(), deadlineBeforeToolTimeout)
	t.Cleanup(cancel)

	delivery, err := clipboard.NewAuto(options).Copy(ctx, unicodeText)

	require.NoError(t, err)
	assert.Equal(t, domain.CopyDeliverySentToTerminal, delivery)
	assert.Contains(t, logs.String(), fallbackLogMessage)
}

func TestAuto_Copy_Concurrent(t *testing.T) {
	t.Parallel()

	tools := newFakeTools(t)
	tools.install(t, "pbcopy", recordingScript)
	auto := clipboard.NewAuto(tools.options(darwinGOOS))

	for copyNumber := range concurrentCopies {
		t.Run(fmt.Sprintf("copy %d", copyNumber), func(t *testing.T) {
			t.Parallel()

			delivery, err := auto.Copy(t.Context(), unicodeText)

			require.NoError(t, err)
			assert.Equal(t, domain.CopyDeliveryPlaced, delivery)
		})
	}
}

func withLogBuffer(options clipboard.Options) (clipboard.Options, *bytes.Buffer) {
	var buffer bytes.Buffer

	options.Logger = slog.New(slog.NewJSONHandler(&buffer, nil))

	return options, &buffer
}
