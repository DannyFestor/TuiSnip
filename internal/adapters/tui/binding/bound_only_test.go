package binding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
)

func TestBoundOnly(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"enter", "down"}, binding.BoundOnly("enter", "", "down"))
}
