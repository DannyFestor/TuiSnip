package pointer

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"

const frameInset = look.BorderWidth / 2

type Point struct {
	X int
	Y int
}

func (p Point) Relative(origin Point) Point {
	return Point{X: p.X - origin.X, Y: p.Y - origin.Y}
}

func (p Point) Within(box look.Size) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < box.Width && p.Y < box.Height
}

func (p Point) InsideFrame() Point {
	return p.Relative(Point{X: frameInset, Y: frameInset})
}
