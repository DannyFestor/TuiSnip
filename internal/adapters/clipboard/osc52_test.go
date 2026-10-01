package clipboard_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/clipboard"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestOSC52_Copy(t *testing.T) {
	t.Parallel()

	delivery, err := clipboard.NewOSC52().Copy(t.Context(), unicodeText)

	require.NoError(t, err)
	assert.Equal(t, domain.CopyDeliverySentToTerminal, delivery)
}
