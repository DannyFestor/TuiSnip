package tabstop_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
)

func TestExpandFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		column int
		want   string
	}{
		{name: "keeps text without tabs", text: "a界", column: 0, want: "a界"},
		{name: "draws a tab as the spaces up to the next stop", text: "a\tb", column: 0, want: "a   b"},
		{name: "draws a tab after a wide rune up to the next stop", text: "界\tb", column: 0, want: "界  b"},
		{name: "starts the tab stops from the given column", text: "\tb", column: 3, want: " b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tabstop.ExpandFrom([]rune(tt.text), tt.column))
		})
	}
}

func TestExpand(t *testing.T) {
	t.Parallel()

	got := tabstop.Expand("a\tb\n\tc\nab\t\td")

	assert.Equal(t, "a   b\n    c\nab      d", got)
}
