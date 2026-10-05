package value_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestLanguage_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "accepts a chroma name", raw: "Go", want: "Go"},
		{name: "accepts a name with a space", raw: "Bash Session", want: "Bash Session"},
		{name: "accepts plaintext", raw: "plaintext", want: "plaintext"},
		{name: "rejects a different case", raw: "go", wantErr: value.ErrUnknownLanguage},
		{name: "rejects an alias", raw: "golang", wantErr: value.ErrUnknownLanguage},
		{name: "rejects surrounding space", raw: " Go", wantErr: value.ErrUnknownLanguage},
		{name: "rejects empty name", raw: "", wantErr: value.ErrUnknownLanguage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewLanguage(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestLanguages(t *testing.T) {
	t.Parallel()

	languages := value.Languages()

	assert.Contains(t, languages, value.PlainText())

	for _, language := range languages {
		parsed, err := value.NewLanguage(language.String())
		require.NoError(t, err)
		assert.Equal(t, language, parsed)
	}
}

func TestLanguage_PlainText(t *testing.T) {
	t.Parallel()

	parsed, err := value.NewLanguage(value.PlainText().String())

	require.NoError(t, err)
	assert.Equal(t, value.PlainText(), parsed)
}
