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
		msgs []tea.Msg
		want string
	}{
		{name: "down moves to the next row", msgs: []tea.Msg{keypress.Letter('j')}, want: "row 2"},
		{name: "G goes to the last row", msgs: []tea.Msg{keypress.Letter('G')}, want: "row 5"},
		{name: "page down moves by the box height", msgs: []tea.Msg{keypress.Special(tea.KeyPgDown)}, want: "row 4"},
		{name: "ignores keys without a Binding", msgs: []tea.Msg{keypress.Letter('x')}, want: "row 1"},
		{
			name: "a click selects the clicked row",
			msgs: []tea.Msg{pointer.Clicked{At: pointer.Point{X: 0, Y: 2}, Double: false}},
			want: "row 3",
		},
		{
			name: "a click below the last visible row does nothing",
			msgs: []tea.Msg{pointer.Clicked{At: pointer.Point{X: 0, Y: boxHeight}, Double: false}},
			want: "row 1",
		},
		{
			name: "the wheel scrolls without moving the cursor",
			msgs: []tea.Msg{pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 2}},
			want: "row 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := listOf(t, numberedRows(5)...)
			for _, msg := range tt.msgs {
				list = list.Update(msg)
			}

			assert.Equal(t, tt.want, selected(list))
		})
	}

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

		list := listOf(t, numberedRows(5)...).Update(pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: 2})

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

	assert.True(t, list.HasRowAt(pointer.Point{X: 0, Y: 1}))
	assert.False(t, list.HasRowAt(pointer.Point{X: 0, Y: 2}))
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
