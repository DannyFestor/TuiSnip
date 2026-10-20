package textarea

import (
	"slices"
	"sync"
)

type rowWrap struct {
	source  []rune
	width   int
	wrapped []wrappedLine
}

func (w rowWrap) wraps(source []rune, width int) bool {
	return len(w.wrapped) > 0 && w.width == width && slices.Equal(w.source, source)
}

type rowWraps struct {
	mutex sync.Mutex
	rows  []rowWrap
}

func (r *rowWraps) lookup(row int, source []rune, width int) ([]wrappedLine, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if row >= len(r.rows) || !r.rows[row].wraps(source, width) {
		return nil, false
	}

	return r.rows[row].wrapped, true
}

func (r *rowWraps) store(row int, source []rune, width int, wrapped []wrappedLine) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if row >= len(r.rows) {
		r.rows = append(r.rows, make([]rowWrap, row+1-len(r.rows))...)
	}

	r.rows[row] = rowWrap{source: slices.Clone(source), width: width, wrapped: wrapped}
}
