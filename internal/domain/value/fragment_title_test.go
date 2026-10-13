package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestFragmentTitle_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "trims surrounding space", raw: "  main.go \n", want: "main.go"},
		{name: "keeps internal space", raw: "read me", want: "read me"},
		{name: "accepts no title", raw: "", want: ""},
		{name: "takes a blank title as none", raw: " \t\n", want: ""},
		{name: "accepts 200 runes", raw: strings.Repeat("é", 200), want: strings.Repeat("é", 200)},
		{name: "rejects 201 runes", raw: strings.Repeat("é", 201), wantErr: value.ErrFragmentTitleTooLong},
		{
			name: "measures after trimming",
			raw:  "  " + strings.Repeat("a", 200) + "  ",
			want: strings.Repeat("a", 200),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewFragmentTitle(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestFragmentTitle_IsSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "is set with text", raw: "main.go", want: true},
		{name: "is not set when blank", raw: "  ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			title, err := value.NewFragmentTitle(tt.raw)
			require.NoError(t, err)

			assert.Equal(t, tt.want, title.IsSet())
		})
	}
}
