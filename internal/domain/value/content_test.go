package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const maxContentBytes = 256 * 1024

func TestContent_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "accepts empty content", raw: "", want: ""},
		{name: "keeps whitespace and tabs", raw: "\tif x {\n\t}\n\n", want: "\tif x {\n\t}\n\n"},
		{
			name: "accepts 256 KiB",
			raw:  strings.Repeat("a", maxContentBytes),
			want: strings.Repeat("a", maxContentBytes),
		},
		{
			name:    "rejects one byte over 256 KiB",
			raw:     strings.Repeat("a", maxContentBytes+1),
			wantErr: value.ErrContentTooLong,
		},
		{
			name:    "counts bytes, not runes",
			raw:     strings.Repeat("é", maxContentBytes/2+1),
			wantErr: value.ErrContentTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewContent(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
