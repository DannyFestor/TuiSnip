package tabstop_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
)

func TestWidthFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		column int
		want   int
	}{
		{name: "counts text without tabs by its cells", text: "a界", column: 0, want: 3},
		{name: "runs a tab at the line start to cell 4", text: "\t", column: 0, want: 4},
		{name: "runs a tab after one cell to cell 4", text: "a\t", column: 0, want: 4},
		{name: "runs a tab on a tab stop to the next stop", text: "abcd\t", column: 0, want: 8},
		{name: "runs a tab after three cells one cell", text: "abc\tx", column: 0, want: 5},
		{name: "starts the tab stops from the given column", text: "\t", column: 2, want: 2},
		{name: "counts nothing for no text", text: "", column: 3, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tabstop.WidthFrom([]rune(tt.text), tt.column))
		})
	}
}
