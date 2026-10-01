package value_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestSearchQuery_IsBlank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "treats an empty query as blank", raw: "", want: true},
		{name: "treats a query of spaces as blank", raw: "   ", want: true},
		{name: "treats tabs and newlines as blank", raw: "\t\n", want: true},
		{name: "keeps a query with text", raw: "docker", want: false},
		{name: "keeps a query with text inside spaces", raw: " docker ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, value.NewSearchQuery(tt.raw).IsBlank())
		})
	}
}

func TestSearchQuery_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, " docker run ", value.NewSearchQuery(" docker run ").String())
}
