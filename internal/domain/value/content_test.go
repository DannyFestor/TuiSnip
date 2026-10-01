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

func TestContent_WithoutTrailingNewline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "drops one trailing newline", raw: "echo hi\n", want: "echo hi"},
		{name: "drops a trailing CRLF as a unit", raw: "echo hi\r\n", want: "echo hi"},
		{name: "drops only the last of several newlines", raw: "echo hi\n\n", want: "echo hi\n"},
		{name: "keeps content without a trailing newline", raw: "echo hi", want: "echo hi"},
		{name: "keeps a lone trailing carriage return", raw: "echo hi\r", want: "echo hi\r"},
		{name: "keeps empty content", raw: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := value.NewContent(tt.raw)
			require.NoError(t, err)

			assert.Equal(t, tt.want, content.WithoutTrailingNewline().String())
		})
	}
}
