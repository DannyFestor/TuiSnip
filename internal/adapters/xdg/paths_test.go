package xdg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		env     map[string]string
		want    xdg.Paths
		wantErr error
	}{
		{
			name: "uses the XDG directories when set",
			env: map[string]string{
				"HOME":            "/home/ada",
				"XDG_CONFIG_HOME": "/xdg/config",
				"XDG_DATA_HOME":   "/xdg/data",
				"XDG_STATE_HOME":  "/xdg/state",
			},
			want: xdg.Paths{
				ConfigFile:   "/xdg/config/tuisnip/config.toml",
				DatabaseFile: "/xdg/data/tuisnip/tuisnip.db",
				StateFile:    "/xdg/state/tuisnip/state.toml",
				LogFile:      "/xdg/state/tuisnip/tuisnip.log",
			},
		},
		{
			name: "falls back to the home directory when unset",
			env:  map[string]string{"HOME": "/home/ada"},
			want: xdg.Paths{
				ConfigFile:   "/home/ada/.config/tuisnip/config.toml",
				DatabaseFile: "/home/ada/.local/share/tuisnip/tuisnip.db",
				StateFile:    "/home/ada/.local/state/tuisnip/state.toml",
				LogFile:      "/home/ada/.local/state/tuisnip/tuisnip.log",
			},
		},
		{
			name: "ignores relative XDG directories",
			env: map[string]string{
				"HOME":            "/home/ada",
				"XDG_CONFIG_HOME": "config",
				"XDG_DATA_HOME":   "./data",
				"XDG_STATE_HOME":  "state",
			},
			want: xdg.Paths{
				ConfigFile:   "/home/ada/.config/tuisnip/config.toml",
				DatabaseFile: "/home/ada/.local/share/tuisnip/tuisnip.db",
				StateFile:    "/home/ada/.local/state/tuisnip/state.toml",
				LogFile:      "/home/ada/.local/state/tuisnip/tuisnip.log",
			},
		},
		{
			name:    "rejects a missing home directory",
			env:     map[string]string{"XDG_CONFIG_HOME": "/xdg/config"},
			wantErr: xdg.ErrNoHome,
		},
		{
			name:    "rejects a relative home directory",
			env:     map[string]string{"HOME": "home/ada"},
			wantErr: xdg.ErrNoHome,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := xdg.Resolve(func(key string) string { return tt.env[key] })

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
