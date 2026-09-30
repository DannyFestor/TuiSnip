package sqltype_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const storedID = "0192f0c1-7a3b-7c4d-8e5f-6a7b8c9d0e1f"

func TestID_Value(t *testing.T) {
	t.Parallel()

	id := sqltype.ID(uuid.MustParse(storedID))

	got, err := id.Value()

	require.NoError(t, err)
	assert.Equal(t, storedID, got)
}

func TestID_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     any
		wantErr error
	}{
		{name: "reads TEXT as string", src: storedID},
		{name: "reads TEXT as bytes", src: []byte(storedID)},
		{name: "reads uppercase TEXT", src: "0192F0C1-7A3B-7C4D-8E5F-6A7B8C9D0E1F"},
		{name: "rejects malformed TEXT as corrupt", src: "not-a-uuid", wantErr: domain.ErrCorruptRecord},
		{name: "rejects INTEGER as corrupt", src: int64(42), wantErr: domain.ErrCorruptRecord},
		{name: "rejects NULL as corrupt", src: nil, wantErr: domain.ErrCorruptRecord},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got sqltype.ID

			err := got.Scan(tt.src)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, uuid.MustParse(storedID), uuid.UUID(got))
		})
	}
}
