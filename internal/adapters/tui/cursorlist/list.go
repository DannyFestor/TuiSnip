package cursorlist

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

type List[T any] struct {
	global binding.Set
	rows   []T
	cursor move.Cursor
	box    look.Size
}

func New[T any](keys binding.Keys) List[T] {
	return List[T]{
		global: keys.For(binding.ScopeGlobal),
		rows:   nil,
		cursor: move.Cursor{},
		box:    look.Size{Width: 0, Height: 0},
	}
}

func (l List[T]) Update(msg tea.Msg) List[T] {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return l.pressed(msg)
	case look.Resized:
		return l.resized(msg.Box)
	case pointer.Clicked:
		l.cursor, _ = l.cursor.ClickedAt(msg.At.Y, len(l.rows), l.box.Height)
	case pointer.Wheeled:
		l.cursor = l.cursor.Scrolled(msg.Lines, len(l.rows), l.box.Height)
	}

	return l
}

func (l List[T]) View(render func(row T, width int) string, frame look.FrameStyle) string {
	return l.cursor.VisibleRows(len(l.rows), l.box.Height, func(index int) string {
		return render(l.rows[index], l.box.Width)
	}, frame.CursorOn)
}

func (l List[T]) WithRows(rows []T) List[T] {
	next := l
	next.rows = rows

	return next.withCursor(l.cursor.Index())
}

func (l List[T]) WithCursorOn(matches func(T) bool) List[T] {
	index := slices.IndexFunc(l.rows, matches)
	if index < 0 {
		return l
	}

	return l.withCursor(index)
}

func (l List[T]) Moved(direction move.Direction) List[T] {
	l.cursor = l.cursor.Moved(direction, len(l.rows), l.box.Height)

	return l
}

func (l List[T]) Rows() []T {
	return l.rows
}

func (l List[T]) Selected() (T, bool) {
	return l.rowAt(l.cursor.Index())
}

func (l List[T]) Successor() (T, bool) {
	index := l.cursor.Index() + 1
	if index == len(l.rows) {
		index = l.cursor.Index() - 1
	}

	return l.rowAt(index)
}

func (l List[T]) HasRowAt(at pointer.Point) bool {
	_, ok := l.cursor.RowAt(at.Y, len(l.rows), l.box.Height)

	return ok
}

func (l List[T]) pressed(msg tea.KeyPressMsg) List[T] {
	direction, ok := move.Pressed(l.global, msg)
	if !ok {
		return l
	}

	return l.Moved(direction)
}

func (l List[T]) resized(box look.Size) List[T] {
	if box == l.box {
		return l
	}

	next := l
	next.box = box

	return next.withCursor(l.cursor.Index())
}

func (l List[T]) withCursor(index int) List[T] {
	l.cursor = l.cursor.At(index, len(l.rows), l.box.Height)

	return l
}

func (l List[T]) rowAt(index int) (T, bool) {
	if index < 0 || index >= len(l.rows) {
		var none T

		return none, false
	}

	return l.rows[index], true
}
