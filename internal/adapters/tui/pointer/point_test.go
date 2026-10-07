package pointer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestPoint_Within(t *testing.T) {
	t.Parallel()

	box := look.Size{Width: 4, Height: 3}

	tests := []struct {
		name  string
		point pointer.Point
		want  bool
	}{
		{name: "holds the first cell", point: pointer.Point{X: 0, Y: 0}, want: true},
		{name: "holds the last cell", point: pointer.Point{X: 3, Y: 2}, want: true},
		{name: "leaves out the column past the right edge", point: pointer.Point{X: 4, Y: 0}, want: false},
		{name: "leaves out the row past the bottom edge", point: pointer.Point{X: 0, Y: 3}, want: false},
		{name: "leaves out a cell left of the box", point: pointer.Point{X: -1, Y: 0}, want: false},
		{name: "leaves out a cell above the box", point: pointer.Point{X: 0, Y: -1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.point.Within(box))
		})
	}
}

func TestPoint_Relative(t *testing.T) {
	t.Parallel()

	got := pointer.Point{X: 5, Y: 9}.Relative(pointer.Point{X: 2, Y: 4})

	assert.Equal(t, pointer.Point{X: 3, Y: 5}, got)
}

func TestPoint_InsideFrame(t *testing.T) {
	t.Parallel()

	assert.Equal(t, pointer.Point{X: 0, Y: 0}, pointer.Point{X: 1, Y: 1}.InsideFrame())
}

func TestClicked_InsideFrame(t *testing.T) {
	t.Parallel()

	got := pointer.Clicked{At: pointer.Point{X: 5, Y: 9}, Double: true}.InsideFrame()

	assert.Equal(t, pointer.Clicked{At: pointer.Point{X: 4, Y: 8}, Double: true}, got)
}
