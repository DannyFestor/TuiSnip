package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestDescription_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "accepts empty description", raw: "", want: ""},
		{name: "keeps surrounding space", raw: "  for APIs \n", want: "  for APIs \n"},
		{name: "accepts 2000 runes", raw: strings.Repeat("é", 2000), want: strings.Repeat("é", 2000)},
		{name: "rejects 2001 runes", raw: strings.Repeat("é", 2001), wantErr: value.ErrDescriptionTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewDescription(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
