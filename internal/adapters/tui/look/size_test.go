package look_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestSizeOf(t *testing.T) {
	t.Parallel()

	assert.Equal(t, look.Size{Width: 120, Height: 40}, look.SizeOf(tea.WindowSizeMsg{Width: 120, Height: 40}))
}

func TestSize_Inner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		outer look.Size
		want  look.Size
	}{
		{
			name:  "takes the border off each side",
			outer: look.Size{Width: 10, Height: 5},
			want:  look.Size{Width: 8, Height: 3},
		},
		{
			name:  "leaves nothing inside a bare border",
			outer: look.Size{Width: 2, Height: 2},
			want:  look.Size{Width: 0, Height: 0},
		},
		{name: "never goes below zero", outer: look.Size{Width: 1, Height: 0}, want: look.Size{Width: 0, Height: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.outer.Inner())
		})
	}
}

func TestSize_Share(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		want    look.Size
	}{
		{name: "takes the given percent of each side", percent: 80, want: look.Size{Width: 96, Height: 32}},
		{name: "rounds down", percent: 33, want: look.Size{Width: 39, Height: 13}},
		{name: "keeps the whole box at 100", percent: 100, want: look.Size{Width: 120, Height: 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, look.Size{Width: 120, Height: 40}.Share(tt.percent))
		})
	}
}
