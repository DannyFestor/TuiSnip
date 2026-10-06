//go:build feature

package feature_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const plainText = "plaintext"

func create(t *testing.T, app *bootstrap.App, in snippet.CreateInput) domain.Snippet {
	t.Helper()

	created, err := app.Create.Run(t.Context(), in)
	require.NoError(t, err)

	return created
}
