package config_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
)

func TestParseKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		written string
		want    string
	}{
		{name: "accepts a lower-case letter", written: "y", want: "y"},
		{name: "accepts a capital", written: "E", want: "E"},
		{name: "accepts punctuation", written: "?", want: "?"},
		{name: "accepts a plus sign", written: "+", want: "+"},
		{name: "accepts a non-ASCII character", written: "é", want: "é"},
		{name: "accepts a named key", written: "pgdown", want: "pgdown"},
		{name: "accepts space by name", written: "space", want: "space"},
		{name: "accepts the last function key", written: "f12", want: "f12"},
		{name: "accepts a modifier", written: "ctrl+s", want: "ctrl+s"},
		{name: "accepts a modified plus sign", written: "ctrl++", want: "ctrl++"},
		{name: "accepts shift with a named key", written: "shift+tab", want: "shift+tab"},
		{name: "accepts shift with space", written: "shift+space", want: "shift+space"},
		{name: "accepts shift spelled out beside ctrl", written: "ctrl+shift+e", want: "ctrl+shift+e"},
		{name: "reorders modifiers as Bubble Tea writes them", written: "shift+ctrl+up", want: "ctrl+shift+up"},
		{
			name:    "reorders every modifier",
			written: "super+hyper+meta+shift+alt+ctrl+a",
			want:    "ctrl+alt+shift+meta+hyper+super+a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.ParseKey(tt.written)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestParseKey_Rejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		written string
		wantErr string
	}{
		{name: "rejects an empty string", written: "", wantErr: `"" is not a key`},
		{name: "rejects a literal space", written: " ", wantErr: `" " never matches; write "space"`},
		{name: "rejects escape", written: "escape", wantErr: `"escape" never matches; write "esc"`},
		{
			name:    "rejects escape after a modifier",
			written: "alt+escape",
			wantErr: `"alt+escape" never matches; write "alt+esc"`,
		},
		{name: "rejects a shifted letter", written: "shift+e", wantErr: `"shift+e" never matches; write "E"`},
		{name: "rejects a shifted capital", written: "shift+E", wantErr: `"shift+E" never matches; write "E"`},
		{
			name:    "rejects a capital beside ctrl",
			written: "ctrl+E",
			wantErr: `"ctrl+E" never matches; write "ctrl+shift+e"`,
		},
		{
			name:    "rejects a shifted symbol",
			written: "shift+1",
			wantErr: `"shift+1" never matches; write the character it types`,
		},
		{name: "rejects an unknown name", written: "f13", wantErr: `"f13" is not a key`},
		{name: "rejects two characters", written: "ab", wantErr: `"ab" is not a key`},
		{name: "rejects a control character", written: "\t", wantErr: `"\t" is not a key`},
		{name: "rejects invalid UTF-8", written: "\xff", wantErr: `"\xff" is not a key`},
		{name: "rejects a capitalised modifier", written: "Ctrl+s", wantErr: `"Ctrl+s" is not a key`},
		{name: "rejects a repeated modifier", written: "ctrl+ctrl+s", wantErr: `"ctrl+ctrl+s" is not a key`},
		{name: "rejects a bare modifier", written: "ctrl", wantErr: `"ctrl" is not a key`},
		{name: "rejects a trailing separator", written: "ctrl+", wantErr: `"ctrl+" is not a key`},
		{name: "rejects a leading separator", written: "+a", wantErr: `"+a" is not a key`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.ParseKey(tt.written)

			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func FuzzParseKey(f *testing.F) {
	for _, seed := range []string{"", " ", "y", "E", "+", "ctrl++", "shift+ctrl+up", "shift+e", "ctrl+E", "f12", "+a"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, written string) {
		key, err := config.ParseKey(written)
		if err != nil {
			assert.True(t, strings.HasPrefix(err.Error(), strconv.Quote(written)), "error %q", err)

			return
		}

		reparsed, err := config.ParseKey(key.String())
		require.NoError(t, err)
		assert.Equal(t, key, reparsed)
	})
}
