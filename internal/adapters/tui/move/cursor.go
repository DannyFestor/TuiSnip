package move

import "strings"

type Cursor struct {
	index  int
	offset int
}

func (c Cursor) Index() int {
	return c.index
}

func (c Cursor) VisibleRows(rows, height int, row func(index int) string, highlight func(string) string) string {
	end := max(c.offset, min(rows, c.offset+height))
	lines := make([]string, 0, end-c.offset)

	for index := c.offset; index < end; index++ {
		line := row(index)
		if index == c.index {
			line = highlight(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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
