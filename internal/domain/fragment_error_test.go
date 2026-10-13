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

func TestFragmentError_InFragment(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for a nil error", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, domain.InFragment(1, nil))
	})

	t.Run("names the Fragment's position and wraps the rule", func(t *testing.T) {
		t.Parallel()

		err := domain.InFragment(1, domain.OnField(domain.FieldFragmentTitle, value.ErrFragmentTitleTooLong))

		require.ErrorIs(t, err, value.ErrFragmentTitleTooLong)
		assert.EqualError(t, err, "fragment 2: fragment_title: value: fragment title is longer than 200 characters")
	})

	t.Run("is found under a wrap, and its field errors with it", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("snippet.Update: %w", errors.Join(
			domain.OnField(domain.FieldTitle, value.ErrBlankTitle),
			domain.InFragment(2, domain.OnField(domain.FieldContent, value.ErrContentTooLong)),
		))

		fragmentErr, ok := errors.AsType[domain.FragmentError](err)

		require.True(t, ok)
		assert.Equal(t, 2, fragmentErr.Position)
		assert.Len(t, domain.FieldErrors(err), 2)
	})
}
