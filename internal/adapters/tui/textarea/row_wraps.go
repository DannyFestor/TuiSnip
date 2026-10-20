package textarea

import "slices"

type rowWrap struct {
	source  []rune
	width   int
	wrapped []wrappedLine
}

func (w rowWrap) wraps(source []rune, width int) bool {
	return len(w.wrapped) > 0 && w.width == width && slices.Equal(w.source, source)
}

type rowWraps struct {
	rows []rowWrap
}

func (r *rowWraps) lookup(row int, source []rune, width int) ([]wrappedLine, bool) {
	if row >= len(r.rows) || !r.rows[row].wraps(source, width) {
		return nil, false
	}

	return r.rows[row].wrapped, true
}

func (r *rowWraps) store(row int, source []rune, width int, wrapped []wrappedLine) {
	if row >= len(r.rows) {
		r.rows = append(r.rows, make([]rowWrap, row+1-len(r.rows))...)
	}

	r.rows[row] = rowWrap{source: slices.Clone(source), width: width, wrapped: wrapped}
}
