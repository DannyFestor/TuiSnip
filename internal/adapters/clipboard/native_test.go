package clipboard_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const deadlineBeforeToolTimeout = 50 * time.Millisecond

func TestNative_Copy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goos      string
		env       []string
		installed []string
		wantTool  string
		wantArgs  string
	}{
		{name: "uses pbcopy on macOS", goos: darwinGOOS, installed: []string{"pbcopy", "xclip"}, wantTool: "pbcopy"},
		{
			name: "uses wl-copy on Wayland", goos: linuxGOOS, env: []string{"WAYLAND_DISPLAY=wayland-0", "DISPLAY=:0"},
			installed: []string{"wl-copy", "xclip"}, wantTool: "wl-copy",
		},
		{
			name: "falls back to xclip under XWayland without wl-copy", goos: linuxGOOS,
			env: []string{"WAYLAND_DISPLAY=wayland-0", "DISPLAY=:0"}, installed: []string{"xclip"},
			wantTool: "xclip", wantArgs: "-in -selection clipboard",
		},
		{
			name: "uses xsel on X11 without xclip", goos: linuxGOOS, env: []string{"DISPLAY=:0"},
			installed: []string{"xsel"}, wantTool: "xsel", wantArgs: "--input --clipboard",
		},
		{
			name: "ignores wl-copy without a Wayland display", goos: linuxGOOS, env: []string{"DISPLAY=:0"},
			installed: []string{"wl-copy", "xclip"}, wantTool: "xclip", wantArgs: "-in -selection clipboard",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := newFakeTools(t)
			for _, name := range tt.installed {
				tools.install(t, name, recordingScript)
			}

			delivery, err := clipboard.NewNative(tools.options(tt.goos, tt.env...)).Copy(t.Context(), unicodeText)

			require.NoError(t, err)
			assert.Equal(t, domain.CopyDeliveryPlaced, delivery)
			assert.Equal(t, unicodeText, tools.received(t, tt.wantTool))
			assert.Equal(t, tt.wantArgs, tools.recorded(t, tt.wantTool, ".args"))
		})
	}
}

func TestNative_Copy_Locale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		goos        string
		tool        string
		env         []string
		wantLCCtype string
	}{
		{name: "gives pbcopy a UTF-8 locale when none is set", goos: darwinGOOS, tool: "pbcopy", wantLCCtype: "UTF-8"},
		{
			name: "keeps the locale pbcopy inherits from LANG",
			goos: darwinGOOS,
			tool: "pbcopy",
			env:  []string{"LANG=de_DE.UTF-8"},
		},
		{
			name: "keeps the locale pbcopy inherits from LC_ALL",
			goos: darwinGOOS,
			tool: "pbcopy",
			env:  []string{"LC_ALL=C.UTF-8"},
		},
		{name: "leaves the locale of Linux tools alone", goos: linuxGOOS, tool: "xclip", env: []string{"DISPLAY=:0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := newFakeTools(t)
			tools.install(t, tt.tool, recordingScript)

			_, err := clipboard.NewNative(tools.options(tt.goos, tt.env...)).Copy(t.Context(), unicodeText)

			require.NoError(t, err)
			assert.Equal(t, tt.wantLCCtype, tools.recorded(t, tt.tool, ".lc_ctype"))
		})
	}
}

func TestNative_Copy_Failures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goos      string
		env       []string
		installed map[string]string
		wantErr   error
	}{
		{name: "reports no tool on macOS without pbcopy", goos: darwinGOOS, wantErr: domain.ErrNoClipboardTool},
		{
			name: "reports no tool on Linux without a display",
			goos: linuxGOOS,
			installed: map[string]string{
				"wl-copy": recordingScript,
				"xclip":   recordingScript,
			},
			wantErr: domain.ErrNoClipboardTool,
		},
		{
			name: "reports a timeout when the tool hangs", goos: linuxGOOS, env: []string{"WAYLAND_DISPLAY=wayland-0"},
			installed: map[string]string{"wl-copy": hangingScript}, wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := newFakeTools(t)
			for name, script := range tt.installed {
				tools.install(t, name, script)
			}

			ctx, cancel := context.WithTimeout(t.Context(), deadlineBeforeToolTimeout)
			t.Cleanup(cancel)

			_, err := clipboard.NewNative(tools.options(tt.goos, tt.env...)).Copy(ctx, unicodeText)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNative_Copy_ToolExitsNonZero(t *testing.T) {
	t.Parallel()

	tools := newFakeTools(t)
	tools.install(t, "pbcopy", failingScript)

	_, err := clipboard.NewNative(tools.options(darwinGOOS)).Copy(t.Context(), unicodeText)

	require.Error(t, err)
	require.NotErrorIs(t, err, domain.ErrNoClipboardTool)
	assert.ErrorContains(t, err, "pbcopy")
}
