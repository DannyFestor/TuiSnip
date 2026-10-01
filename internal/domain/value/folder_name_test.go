package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestFolderName_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "trims surrounding space", raw: "  docker \n", want: "docker"},
		{name: "keeps internal space", raw: "docker  compose", want: "docker  compose"},
		{name: "rejects empty name", raw: "", wantErr: value.ErrBlankFolderName},
		{name: "rejects blank name", raw: " \t\n", wantErr: value.ErrBlankFolderName},
		{name: "accepts 200 runes", raw: strings.Repeat("é", 200), want: strings.Repeat("é", 200)},
		{name: "rejects 201 runes", raw: strings.Repeat("é", 201), wantErr: value.ErrFolderNameTooLong},
		{
			name: "measures after trimming",
			raw:  "  " + strings.Repeat("a", 200) + "  ",
			want: strings.Repeat("a", 200),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewFolderName(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
