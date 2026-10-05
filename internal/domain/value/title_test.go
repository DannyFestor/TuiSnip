package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestTitle_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "trims surrounding space", raw: "  curl json \n", want: "curl json"},
		{name: "keeps internal space", raw: "curl  json", want: "curl  json"},
		{name: "rejects empty title", raw: "", wantErr: value.ErrBlankTitle},
		{name: "rejects blank title", raw: " \t\n", wantErr: value.ErrBlankTitle},
		{name: "accepts 200 runes", raw: strings.Repeat("é", 200), want: strings.Repeat("é", 200)},
		{name: "rejects 201 runes", raw: strings.Repeat("é", 201), wantErr: value.ErrTitleTooLong},
		{
			name: "measures after trimming",
			raw:  "  " + strings.Repeat("a", 200) + "  ",
			want: strings.Repeat("a", 200),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewTitle(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestTitle_CompareIgnoringCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		other string
		want  int
	}{
		{name: "orders a lowercase title before a later capitalised one", title: "awk", other: "Bash", want: -1},
		{name: "orders a capitalised title after an earlier lowercase one", title: "Bash", other: "awk", want: 1},
		{name: "finds titles differing only in ASCII case equal", title: "Curl", other: "cURL", want: 0},
		{name: "folds a capital A, the first ASCII capital", title: "Awk", other: "awk", want: 0},
		{name: "folds to lowercase, so an underscore orders before a capital", title: "_x", other: "Zx", want: -1},
		{name: "keeps non-ASCII letters unfolded", title: "Écrire", other: "écrire", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, mustTitle(t, tt.title).CompareIgnoringCase(mustTitle(t, tt.other)))
		})
	}
}

func mustTitle(t *testing.T, raw string) value.Title {
	t.Helper()

	title, err := value.NewTitle(raw)
	require.NoError(t, err)

	return title
}
