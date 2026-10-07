package pointer_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestWheeledBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		button tea.MouseButton
		want   pointer.Wheeled
		wantOK bool
	}{
		{
			name:   "turning down scrolls down",
			button: tea.MouseWheelDown,
			want:   pointer.Wheeled{At: pointer.Point{X: 2, Y: 7}, Lines: 3},
			wantOK: true,
		},
		{
			name:   "turning up scrolls up",
			button: tea.MouseWheelUp,
			want:   pointer.Wheeled{At: pointer.Point{X: 2, Y: 7}, Lines: -3},
			wantOK: true,
		},
		{
			name:   "pushing sideways scrolls nothing",
			button: tea.MouseWheelLeft,
			want:   pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 0},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := pointer.WheeledBy(tea.MouseWheelMsg{X: 2, Y: 7, Button: tt.button, Mod: 0})

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWheeled_Relative(t *testing.T) {
	t.Parallel()

	got := pointer.Wheeled{At: pointer.Point{X: 5, Y: 9}, Lines: 3}.Relative(pointer.Point{X: 2, Y: 4})

	assert.Equal(t, pointer.Wheeled{At: pointer.Point{X: 3, Y: 5}, Lines: 3}, got)
}
