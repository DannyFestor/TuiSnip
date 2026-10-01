package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestFieldError_OnField(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for a nil error", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, domain.OnField(domain.FieldTitle, nil))
	})

	t.Run("names the field and wraps the rule", func(t *testing.T) {
		t.Parallel()

		err := domain.OnField(domain.FieldTitle, value.ErrBlankTitle)

		require.ErrorIs(t, err, value.ErrBlankTitle)
		assert.EqualError(t, err, "title: value: title is blank")
	})
}

func TestFieldError_FieldErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want []domain.Field
	}{
		{name: "finds nothing in nil", err: nil, want: nil},
		{name: "finds nothing in a plain error", err: domain.ErrNilID, want: nil},
		{
			name: "finds every joined field",
			err: errors.Join(
				domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
				domain.OnField(domain.FieldDescription, nil),
				domain.OnField(domain.FieldLanguage, value.ErrUnknownLanguage),
			),
			want: []domain.Field{domain.FieldTitle, domain.FieldLanguage},
		},
		{
			name: "finds fields under a wrap",
			err:  fmt.Errorf("snippet.Create: %w", domain.OnField(domain.FieldContent, value.ErrContentTooLong)),
			want: []domain.Field{domain.FieldContent},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []domain.Field
			for _, fieldErr := range domain.FieldErrors(tt.err) {
				got = append(got, fieldErr.Field)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
