package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireErrors(t *testing.T, err error, want []error) {
	t.Helper()

	if len(want) == 0 {
		require.NoError(t, err)

		return
	}

	for _, sentinel := range want {
		require.ErrorIs(t, err, sentinel)
	}
}
