package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestRequireDependency(t *testing.T) {
	t.Parallel()

	t.Run("names a nil dependency", func(t *testing.T) {
		t.Parallel()

		err := domain.RequireDependency("clock", nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		assert.Equal(t, "clock: domain: dependency is missing", err.Error())
	})

	t.Run("accepts a present dependency", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, domain.RequireDependency("clock", struct{}{}))
	})
}
