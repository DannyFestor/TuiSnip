package textarea

import (
	"strconv"
	"strings"
)

type runeSpan struct {
	from int
	to   int
}

// PositionAt maps coordinates within the textarea's rendered area to a buffer
// position. Coordinates are relative to the textarea itself: (0, 0) is its
// top-left cell, including the prompt/line-number gutter. Callers that render
// the textarea at an offset must subtract that offset first.
//
// Coordinates outside the content resolve to the nearest position: above the
// first line yields the start of the buffer, below the last yields its end,
// and a column past a line's text yields the end of that line.
func (m *Model) PositionAt(cellX, cellY int) Position {
	if len(m.value) == 0 {
		return Position{Row: 0, Col: 0}
	}

	targetLine := cellY + m.viewport.YOffset()
	if targetLine < 0 {
		return Position{Row: 0, Col: 0}
	}

	contentX := max(cellX-m.gutterWidth(), 0)

	display := 0

	for row, line := range m.value {
		base := 0

		for _, wrapped := range m.memoizedWrap(line, m.width) {
			if display == targetLine {
				col := base + wrapped.runeIndexAt(contentX)

				return Position{Row: row, Col: clamp(col, 0, len(line))}
			}

			base += len(wrapped.runes)
			display++
		}
	}

	lastRow := len(m.value) - 1

	return Position{Row: lastRow, Col: len(m.value[lastRow])}
}

// BeginSelection starts a selection at the given textarea-relative
// coordinates, discarding any previous selection, and moves the cursor there.
// Pair it with [Model.ExtendSelection] as the pointer moves and
// [Model.EndSelection] when the drag finishes.
//
// See [Model.PositionAt] for the coordinate convention.
func (m *Model) BeginSelection(x, y int) {
	pos := m.PositionAt(x, y)
	m.selectFrom(pos, pos)
	m.selecting = true
	m.moveCursorTo(pos)
}

// ExtendSelection extends an in-progress selection to the given
// textarea-relative coordinates and moves the cursor there. It is a no-op
// unless [Model.BeginSelection] started a drag.
func (m *Model) ExtendSelection(x, y int) {
	if !m.selecting {
		return
	}

	pos := m.PositionAt(x, y)
	m.selHead = pos
	m.hasSelection = true
	m.moveCursorTo(pos)
}

// EndSelection completes an in-progress drag. The selection itself is
// retained so it can be read with [Model.SelectedText]; a zero-width
// selection (a plain click) is discarded.
func (m *Model) EndSelection() {
	m.selecting = false
	if m.selAnchor == m.selHead {
		m.ClearSelection()
	}
}

func (m *Model) SelectAll() {
	if len(m.value) == 0 {
		return
	}

	lastRow := len(m.value) - 1
	m.selectFrom(
		Position{Row: 0, Col: 0},
		Position{Row: lastRow, Col: len(m.value[lastRow])},
	)
	m.selecting = false
}

func (m *Model) ClearSelection() {
	m.hasSelection = false
	m.selecting = false
	m.selAnchor = Position{Row: 0, Col: 0}
	m.selHead = Position{Row: 0, Col: 0}
}

func (m *Model) HasSelection() bool {
	return m.hasSelection && m.selAnchor != m.selHead
}

// Selection returns the selected range, normalized so start sorts before end,
// and whether a non-empty selection is active.
func (m *Model) Selection() (Range, bool) {
	if !m.HasSelection() {
		return Range{Start: Position{Row: 0, Col: 0}, End: Position{Row: 0, Col: 0}}, false
	}

	start, end := m.selAnchor, m.selHead
	if end.before(start) {
		start, end = end, start
	}

	return Range{Start: start, End: end}, true
}

func (m *Model) SelectedText() string {
	selected, ok := m.Selection()
	if !ok {
		return ""
	}

	start, end := selected.Start, selected.End
	if start.Row == end.Row {
		line := m.value[start.Row]

		return string(line[clamp(start.Col, 0, len(line)):clamp(end.Col, 0, len(line))])
	}

	var text strings.Builder

	for row := start.Row; row <= end.Row; row++ {
		line := m.value[row]
		switch row {
		case start.Row:
			text.WriteString(string(line[clamp(start.Col, 0, len(line)):]))
		case end.Row:
			text.WriteString(string(line[:clamp(end.Col, 0, len(line))]))
		default:
			text.WriteString(string(line))
		}

		if row < end.Row {
			text.WriteRune('\n')
		}
	}

	return text.String()
}

// gutterWidth returns the number of columns rendered to the left of the text
// content: the prompt, plus the line-number column when enabled. Mouse
// coordinates are offset by this much before being mapped to a column.
func (m *Model) gutterWidth() int {
	width := m.promptWidth
	if m.ShowLineNumbers {
		width += len(strconv.Itoa(m.MaxHeight)) + lineNumberGap
	}

	return width
}

func (m *Model) selectFrom(anchor, head Position) {
	m.selAnchor = anchor
	m.selHead = head
	m.hasSelection = true
}

func (m *Model) moveCursorTo(pos Position) {
	if len(m.value) == 0 {
		return
	}

	row := clamp(pos.Row, 0, len(m.value)-1)
	m.row = row
	m.col = clamp(pos.Col, 0, len(m.value[row]))
}

// selectionSpanFor returns the half-open rune range [from, to) of the given
// logical row that is selected, restricted to the wrapped segment covering
// [base, base+length). ok is false when the segment holds no selected runes.
func (m *Model) selectionSpanFor(row, base, length int) (runeSpan, bool) {
	noSpan := runeSpan{from: 0, to: 0}

	selected, active := m.Selection()
	if !active || row < selected.Start.Row || row > selected.End.Row {
		return noSpan, false
	}

	lineLen := len(m.value[row])

	rowFrom, rowTo := 0, lineLen
	if row == selected.Start.Row {
		rowFrom = clamp(selected.Start.Col, 0, lineLen)
	}

	if row == selected.End.Row {
		rowTo = clamp(selected.End.Col, 0, lineLen)
	}

	from := max(rowFrom, base)

	to := min(rowTo, base+length)
	if from >= to {
		return noSpan, false
	}

	return runeSpan{from: from - base, to: to - base}, true
}
