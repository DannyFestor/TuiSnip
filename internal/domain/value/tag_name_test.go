package value_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestTagName_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "trims surrounding space", raw: "  go \n", want: "go"},
		{name: "keeps internal space", raw: "shell  tricks", want: "shell  tricks"},
		{name: "keeps the spelling", raw: "GoLang", want: "GoLang"},
		{name: "rejects empty name", raw: "", wantErr: value.ErrBlankTagName},
		{name: "rejects blank name", raw: " \t\n", wantErr: value.ErrBlankTagName},
		{name: "rejects a comma", raw: "go,docker", wantErr: value.ErrTagNameHasComma},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewTagName(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestTagName_Key(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		same bool
	}{
		{name: "ignores ASCII case", a: "Go", b: "go", same: true},
		{name: "ignores case beyond ASCII", a: "Ärger", b: "äRGER", same: true},
		{name: "folds the Kelvin sign with k", a: "Kube", b: "kube", same: true},
		{name: "folds the long s with s", a: "ſhell", b: "Shell", same: true},
		{name: "tells different names apart", a: "go", b: "golang", same: false},
		{name: "keeps diacritics apart", a: "cafe", b: "café", same: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, err := value.NewTagName(tt.a)
			require.NoError(t, err)

			b, err := value.NewTagName(tt.b)
			require.NoError(t, err)

			assert.Equal(t, tt.same, a.Key() == b.Key())
		})
	}
}
