package move

type Cursor struct {
	index  int
	offset int
}

func (c Cursor) Index() int {
	return c.index
}

func (c Cursor) Offset() int {
	return c.offset
}

func (c Cursor) End(rows, height int) int {
	return max(c.offset, min(rows, c.offset+height))
}

func (c Cursor) Moved(direction Direction, rows, height int) Cursor {
	return c.At(c.target(direction, rows, height), rows, height)
}

func (c Cursor) At(index, rows, height int) Cursor {
	c.index = max(0, min(index, rows-1))
	c.offset = max(0, min(c.offset, c.index), c.index-height+1)

	return c
}

func (c Cursor) target(direction Direction, rows, height int) int {
	switch direction {
	case Down:
		return c.index + 1
	case Up:
		return c.index - 1
	case Top:
		return 0
	case Bottom:
		return rows - 1
	case PageDown:
		return c.index + height
	case PageUp:
		return c.index - height
	case None:
		return c.index
	}

	return c.index
}
