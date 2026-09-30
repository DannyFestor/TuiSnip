package sqltype_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const storedNanoseconds int64 = 1_727_740_800_123_456_789

func TestTimestamp_Value(t *testing.T) {
	t.Parallel()

	tokyo := time.FixedZone("JST", 9*60*60)
	timestamp := sqltype.Timestamp(time.Unix(0, storedNanoseconds).In(tokyo))

	got, err := timestamp.Value()

	require.NoError(t, err)
	assert.Equal(t, storedNanoseconds, got)
}

func TestTimestamp_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     any
		wantErr error
	}{
		{name: "reads INTEGER nanoseconds in UTC", src: storedNanoseconds},
		{name: "rejects TEXT as corrupt", src: "2024-10-01", wantErr: domain.ErrCorruptRecord},
		{name: "rejects NULL as corrupt", src: nil, wantErr: domain.ErrCorruptRecord},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got sqltype.Timestamp

			err := got.Scan(tt.src)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, time.Unix(0, storedNanoseconds).UTC(), time.Time(got))
		})
	}
}
