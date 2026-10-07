package pointer

type Clicked struct {
	At     Point
	Double bool
}

func (c Clicked) Relative(origin Point) Clicked {
	c.At = c.At.Relative(origin)

	return c
}

func (c Clicked) InsideFrame() Clicked {
	c.At = c.At.InsideFrame()

	return c
}
