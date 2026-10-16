package value_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestFolderPath_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     []string
		want    []string
		wantErr []error
	}{
		{name: "an empty path", raw: []string{}, want: []string{}},
		{name: "a missing path reads as empty", raw: nil, want: []string{}},
		{name: "keeps the names Root first", raw: []string{"scripts", "a/b"}, want: []string{"scripts", "a/b"}},
		{name: "trims each name", raw: []string{" scripts "}, want: []string{"scripts"}},
		{
			name:    "refuses every invalid name",
			raw:     []string{"", "ok", " "},
			want:    []string{},
			wantErr: []error{value.ErrBlankFolderName},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewFolderPath(tt.raw)

			for _, wantErr := range tt.wantErr {
				require.ErrorIs(t, err, wantErr)
			}

			if len(tt.wantErr) == 0 {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got.Strings())
		})
	}
}

func TestFolderPath_IsEmpty(t *testing.T) {
	t.Parallel()

	empty, err := value.NewFolderPath(nil)
	require.NoError(t, err)

	held, err := value.NewFolderPath([]string{"scripts"})
	require.NoError(t, err)

	assert.True(t, empty.IsEmpty())
	assert.True(t, value.FolderPath{}.IsEmpty())
	assert.False(t, held.IsEmpty())
}
