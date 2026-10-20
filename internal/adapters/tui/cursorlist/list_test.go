package cursorlist_test

import (
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/cursorlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/screencell"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 12
	boxHeight = 3
)

func TestList_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
		want string
	}{
		{name: "down moves to the next row", key: keypress.Letter('j'), want: "row 2"},
		{name: "G goes to the last row", key: keypress.Letter('G'), want: "row 5"},
		{name: "page down moves by the box height", key: keypress.Special(tea.KeyPgDown), want: "row 4"},
		{name: "ignores keys without a Binding", key: keypress.Letter('x'), want: "row 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := listOf(t, numberedRows(5)...).Update(tt.key)

			assert.Equal(t, tt.want, selected(list))
		})
	}

	t.Run("a click selects the clicked row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...)

		list = list.Update(pointer.Clicked{At: cellOf(t, list, "row 3"), Double: false})

		assert.Equal(t, "row 3", selected(list))
	})

	t.Run("a click below the last row does nothing", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(2)...)
		below := cellOf(t, list, "row 2")
		below.Y++

		list = list.Update(pointer.Clicked{At: below, Double: false})

		assert.Equal(t, "row 1", selected(list))
	})

	t.Run("the wheel scrolls without moving the cursor", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...)

		list = list.Update(pointer.Wheeled{At: cellOf(t, list, "row 2"), Lines: 2})

		assert.Equal(t, "row 1", selected(list))
	})

	t.Run("a shorter box keeps the cursor in view", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).Update(keypress.Letter('G'))

		list = list.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: 2}})

		assert.Equal(t, []string{"row 4", "ROW 5"}, shown(list))
	})
}

func TestList_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the rows in view at the box width and marks the cursor", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).Update(keypress.Letter('j'))

		assert.Equal(t, []string{"row 1", "ROW 2", "row 3"}, shown(list))
	})

	t.Run("follows the wheel", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...)

		list = list.Update(pointer.Wheeled{At: cellOf(t, list, "row 2"), Lines: 2})

		assert.Equal(t, []string{"row 3", "row 4", "row 5"}, shown(list))
	})
}

func TestList_WithRows(t *testing.T) {
	t.Parallel()

	t.Run("keeps the cursor's row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).Update(keypress.Letter('j'))

		list = list.WithRows([]string{"a", "b", "c"})

		assert.Equal(t, "b", selected(list))
	})

	t.Run("moves the cursor onto the last row when its row is gone", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).Update(keypress.Letter('G'))

		list = list.WithRows([]string{"a", "b"})

		assert.Equal(t, "b", selected(list))
	})
}

func TestList_WithCursorOn(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the first matching row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).WithCursorOn(func(row string) bool { return row == "row 4" })

		assert.Equal(t, "row 4", selected(list))
	})

	t.Run("keeps the cursor without a matching row", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedRows(5)...).Update(keypress.Letter('j')).
			WithCursorOn(func(row string) bool { return row == "missing" })

		assert.Equal(t, "row 2", selected(list))
	})
}

func TestList_Moved(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedRows(5)...).Update(keypress.Letter('G')).Moved(move.Top)

	assert.Equal(t, "row 1", selected(list))
}

func TestList_Selected(t *testing.T) {
	t.Parallel()

	_, ok := listOf(t).Selected()

	assert.False(t, ok)
}

func TestList_Successor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rows   int
		keys   []tea.KeyPressMsg
		want   string
		wantOK bool
	}{
		{name: "is the row below the cursor", rows: 3, keys: nil, want: "row 2", wantOK: true},
		{
			name:   "is the row above on the last row",
			rows:   3,
			keys:   []tea.KeyPressMsg{keypress.Letter('G')},
			want:   "row 2",
			wantOK: true,
		},
		{name: "is none for a single row", rows: 1, keys: nil, want: "", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := listOf(t, numberedRows(tt.rows)...)
			for _, key := range tt.keys {
				list = list.Update(key)
			}

			got, ok := list.Successor()

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestList_HasRowAt(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedRows(2)...)
	last := cellOf(t, list, "row 2")
	below := pointer.Point{X: last.X, Y: last.Y + 1}

	assert.True(t, list.HasRowAt(last))
	assert.False(t, list.HasRowAt(below))
}

func cellOf(t *testing.T, list cursorlist.List[string], text string) pointer.Point {
	t.Helper()

	return screencell.Find(t, strings.Join(shown(list), "\n"), text)
}

func listOf(t *testing.T, rows ...string) cursorlist.List[string] {
	t.Helper()

	list := cursorlist.New[string](testsettings.Default(t).Keys)

	return list.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: boxHeight}}).WithRows(rows)
}

func numberedRows(count int) []string {
	rows := make([]string, 0, count)
	for number := 1; number <= count; number++ {
		rows = append(rows, "row "+strconv.Itoa(number))
	}

	return rows
}

func selected(list cursorlist.List[string]) string {
	row, _ := list.Selected()

	return row
}

func shown(list cursorlist.List[string]) []string {
	frame := look.FrameStyle{
		Border: lipgloss.NewStyle(),
		Title:  lipgloss.NewStyle(),
		Cursor: lipgloss.NewStyle().Transform(strings.ToUpper),
	}

	view := list.View(func(row string, width int) string { return look.Row(row, "", width) }, frame)

	lines := strings.Split(ansi.Strip(view), "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " ")
	}

	return lines
}
