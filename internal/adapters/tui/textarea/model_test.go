package textarea_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestVerticalScrolling(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.Prompt = ""
	model.ShowLineNumbers = false
	model.SetHeight(1)
	model.SetWidth(20)
	model.CharLimit = 100

	model, _ = model.Update(nil)

	input := "This is a really long line that should wrap around the text area."

	model = sendString(model, input)

	view := model.View()

	// The view should contain the end of "line" of the input.
	assert.Contains(t, view, "the text area.", "Text area did not render the input")

	// But we should be able to scroll to see the next line.
	// Let's scroll down for each line to view the full input.
	lines := []string{
		"This is a really",
		"long line that",
		"should wrap around",
		"the text area.",
	}

	model.MoveToBegin()

	for _, line := range lines {
		view = ansi.Strip(model.View())
		assert.Contains(t, view, line, "Text area did not render the correct scrolled input")

		model.CursorDown()
	}
}

func TestWordWrapOverflowing(t *testing.T) {
	t.Parallel()

	// An interesting edge case is when the user enters many words that fill up
	// the text area and then goes back up and inserts a few words which causes
	// a cascading wrap and causes an overflow of the last line.
	//
	// In this case, we should not let the user insert more words if, after the
	// entire wrap is complete, the last line is overflowing.
	model := newTextArea()

	model.SetHeight(3)
	model.SetWidth(20)
	model.CharLimit = 500

	model, _ = model.Update(nil)

	input := strings.TrimSuffix(strings.Repeat("Testing ", 8), " ")
	model = sendStringViewing(model, input)

	// We have essentially filled the text area with input.
	// Let's see if we can cause wrapping to overflow the last line.
	model.MoveToBegin()

	model = sendStringViewing(model, "Testing")

	assert.LessOrEqual(t, model.LineInfo().Width, 20, model.View())
}

func TestValueSoftWrap(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.SetWidth(16)
	model.SetHeight(10)
	model.CharLimit = 500

	model, _ = model.Update(nil)

	input := strings.TrimSuffix(strings.Repeat("Testing ", 8), " ")
	model = sendStringViewing(model, input)

	require.Equal(t, input, model.Value(), "The text area does not have the correct value")
}

func TestSetValue(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.SetValue(strings.Join([]string{"Foo", "Bar", "Baz"}, "\n"))

	require.Equal(t, 2, model.Line(), "Cursor Should be on row 2 column 3 after inserting 2 new lines")
	require.Equal(t, 3, model.Column(), "Cursor Should be on row 2 column 3 after inserting 2 new lines")
	require.Equal(t, "Foo\nBar\nBaz", model.Value())

	// SetValue should reset text area
	model.SetValue("Test")

	require.Equal(t, "Test", model.Value(), "Text area was not reset when SetValue() was called")
}

func TestInsertString(t *testing.T) {
	t.Parallel()

	model := newTextArea()

	// Insert some text
	model = sendString(model, "foo baz")

	// Put cursor in the middle of the text
	model.SetCursorColumn(4)

	model.InsertString("bar ")

	require.Equal(t, "foo bar baz", model.Value(), "Expected insert string to insert bar between foo and baz")
}

func TestCanHandleEmoji(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	input := "🧋"

	model = sendString(model, input)

	require.Equal(t, input, model.Value(), "Expected emoji to be inserted")

	input = "🧋🧋🧋"

	model.SetValue(input)

	require.Equal(t, input, model.Value(), "Expected emoji to be inserted")
	require.Equal(t, 3, model.Column(), "Expected cursor to be on the third character")
	require.Equal(t, 6, model.LineInfo().CharOffset, "Expected cursor to be on the sixth character")
}

func TestVerticalNavigationKeepsCursorHorizontalPosition(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.SetWidth(20)

	model.SetValue("你好你好" + "\n" + "Hello")

	model = placeCursor(model, 0, 2)

	// 你好|你好
	// Hell|o
	// 1234|

	// Let's imagine our cursor is on the first line where the pipe is.
	// We press the down arrow to get to the next line.
	// The issue is that if we keep the cursor on the same column, the cursor will jump to after the `e`.
	//
	// 你好|你好
	// He|llo
	//
	// But this is wrong because visually we were at the 4th character due to
	// the first line containing double-width runes.
	// We want to keep the cursor on the same visual column.
	//
	// 你好|你好
	// Hell|o
	//
	// This test ensures that the cursor is kept on the same visual column by
	// ensuring that the column offset goes from 2 -> 4.

	lineInfo := model.LineInfo()
	require.Equal(
		t,
		[]int{4, 2},
		[]int{lineInfo.CharOffset, lineInfo.ColumnOffset},
		"Expected cursor to be on the fourth character because there are two double width runes on the first line.",
	)

	model, _ = model.Update(keypress.Special(tea.KeyDown))

	lineInfo = model.LineInfo()
	require.Equal(
		t,
		[]int{4, 4},
		[]int{lineInfo.CharOffset, lineInfo.ColumnOffset},
		"Expected cursor to be on the fourth character because we came down from the first line.",
	)
}

func TestVerticalNavigationShouldRememberPositionWhileTraversing(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.SetWidth(40)

	// Let's imagine we have a text area with the following content:
	//
	// Hello
	// World
	// This is a long line.
	//
	// If we are at the end of the last line and go up, we should be at the end
	// of the second line.
	// And, if we go up again we should be at the end of the first line.
	// But, if we go back down twice, we should be at the end of the last line
	// again and not the fifth (length of second line) character of the last line.
	//
	// In other words, we should remember the last horizontal position while
	// traversing vertically.

	model.SetValue(strings.Join([]string{"Hello", "World", "This is a long line."}, "\n"))

	// We are at the end of the last line.
	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 2, Col: 20},
		"Expected cursor to be on the 20th character of the last line",
	)

	// Let's go up.
	upMsg := keypress.Special(tea.KeyUp)
	model, _ = model.Update(upMsg)

	// We should be at the end of the second line.
	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 1, Col: 5},
		"Expected cursor to be on the 5th character of the second line",
	)

	// And, again.
	model, _ = model.Update(upMsg)

	// We should be at the end of the first line.
	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 0, Col: 5},
		"Expected cursor to be on the 5th character of the first line",
	)

	// Let's go down, twice.
	downMsg := keypress.Special(tea.KeyDown)
	model, _ = model.Update(downMsg)
	model, _ = model.Update(downMsg)

	// We should be at the end of the last line.
	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 2, Col: 20},
		"Expected cursor to be on the 20th character of the last line",
	)

	// Now, for correct behavior, if we move right or left, we should forget
	// (reset) the saved horizontal position. Since we assume the user wants to
	// keep the cursor where it is horizontally. This is how most text areas
	// work.

	model, _ = model.Update(upMsg)
	model, _ = model.Update(keypress.Special(tea.KeyLeft))

	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 1, Col: 4},
		"Expected cursor to be on the 5th character of the second line",
	)

	// Going down now should keep us at the 4th column since we moved left and
	// reset the horizontal position saved state.
	model, _ = model.Update(downMsg)
	requireCursorAt(
		t,
		model,
		textarea.Position{Row: 2, Col: 4},
		"Expected cursor to be on the 4th character of the last line",
	)
}

func TestView(t *testing.T) {
	t.Parallel()

	type want struct {
		view      string
		cursorRow int
		cursorCol int
	}

	tests := []struct {
		name      string
		modelFunc func(textarea.Model) textarea.Model
		want      want
	}{
		{
			name: "placeholder",
			want: want{
				view: heredoc.Doc(`
					>   1 Hello, World!
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "single line",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line")

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line\nthe second line\nthe third line")

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>   2 the second line
					>   3 the third line
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line without line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line")
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multipline lines without line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> the second line
					> the third line
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line and custom end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line")
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					> *
					> *
					> *
					> *
					> *
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines and custom end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 the first line
					>   2 the second line
					>   3 the third line
					> *
					> *
					> *
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line without line numbers and custom end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line")
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> *
					> *
					> *
					> *
					> *
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines without line numbers and custom end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> the first line
					> the second line
					> the third line
					> *
					> *
					> *
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "single line and custom prompt",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line")
				m.Prompt = "* "

				return m
			},
			want: want{
				view: heredoc.Doc(`
					*   1 the first line
					*
					*
					*
					*
					*
				`),
				cursorRow: 0,
				cursorCol: 14,
			},
		},
		{
			name: "multiple lines and custom prompt",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetValue("the first line\nthe second line\nthe third line")
				m.Prompt = "* "

				return m
			},
			want: want{
				view: heredoc.Doc(`
					*   1 the first line
					*   2 the second line
					*   3 the third line
					*
					*
					*
				`),
				cursorRow: 2,
				cursorCol: 14,
			},
		},
		{
			name: "type single line",
			modelFunc: func(m textarea.Model) textarea.Model {
				input := "foo"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "type multiple lines",
			modelFunc: func(m textarea.Model) textarea.Model {
				input := "foo\nbar\nbaz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo
					>   2 bar
					>   3 baz
					>
					>
					>
				`),
				cursorRow: 2,
				cursorCol: 3,
			},
		},
		{
			name: "softwrap",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.Prompt = ""
				m.SetWidth(5)

				input := "foo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					foo
					bar
					baz



				`),
				cursorRow: 2,
				cursorCol: 3,
			},
		},
		{
			name: "single line character limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.CharLimit = 7

				input := "foo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo bar
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 7,
			},
		},
		{
			name: "multiple lines character limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.CharLimit = 19

				input := "foo bar baz\nfoo bar baz"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 foo bar baz
					>   2 foo bar
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 7,
			},
		},
		{
			name: "set width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(10)

				input := "12"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 2,
			},
		},
		{
			name: "set width max length text minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(10)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width max length text",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(10)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width max length text plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(10)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>     5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width set max width minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width set max width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width set max width plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.MaxWidth = 10
				m.SetWidth(11)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1234
					>     5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width min width minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(6)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1
					>     2
					>     3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(7)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 1
					>     2
					>     3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width no line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.SetWidth(0)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1
					> 2
					> 3
					>
					>
					>
				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width no line numbers no prompt",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.Prompt = ""
				m.SetWidth(0)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					1
					2
					3



				`),
				cursorRow: 3,
				cursorCol: 0,
			},
		},
		{
			name: "set width min width plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(8)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12
					>     3
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width without line numbers max length text minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123
					>
					>
					>
					>
					>
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width without line numbers max length text",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1234
					>
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width without line numbers max length text plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = false
				m.SetWidth(6)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 1234
					> 5
					>
					>
					>
					>
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width with style",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.SetWidth(12)

				input := "1"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1   │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 1,
			},
		},
		{
			name: "set width with style max width minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.SetWidth(12)

				input := "123"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 123 │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 3,
			},
		},
		{
			name: "set width with style max width",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.SetWidth(12)

				input := "1234"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1234│
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width with style max width plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.SetWidth(12)

				input := "12345"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│>   1 1234│
					│>     5   │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "set width without line numbers with style",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "123456"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 123456  │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 6,
			},
		},
		{
			name: "set width without line numbers with style max width minus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "1234567"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 1234567 │
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 0,
				cursorCol: 7,
			},
		},
		{
			name: "set width without line numbers with style max width",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "12345678"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 12345678│
					│>         │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 0,
			},
		},
		{
			name: "set width without line numbers with style max width plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				s := m.Styles()
				s.Focused.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
				m.SetStyles(s)
				m.Focus()

				m.ShowLineNumbers = false
				m.SetWidth(12)

				input := "123456789"
				m = sendString(m, input)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					┌──────────┐
					│> 12345678│
					│> 9       │
					│>         │
					│>         │
					│>         │
					│>         │
					└──────────┘
				`),
				cursorRow: 1,
				cursorCol: 1,
			},
		},
		{
			name: "placeholder min width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetWidth(0)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 H
					>     e
					>     l
					>     l
					>     o
					>     ,
				`),
			},
		},
		{
			name: "placeholder single line",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					>
					>
					>
					>
					>
					`),
			},
		},
		{
			name: "placeholder multiple lines",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> placeholder the second line
					> placeholder the third line
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line with line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = true

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines with line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = true

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>     placeholder the second line
					>     placeholder the third line
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line with end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> *
					> *
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder multiple lines with end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = false
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line
					> placeholder the second line
					> placeholder the third line
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder single line with line numbers and end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line"
				m.ShowLineNumbers = true
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					> *
					> *
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder multiple lines with line numbers and end of buffer character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line\nplaceholder the second line\nplaceholder the third line"
				m.ShowLineNumbers = true
				m.EndOfBufferCharacter = '*'

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line
					>     placeholder the second line
					>     placeholder the third line
					> *
					> *
					> *
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line that is longer than the max width"
				m.SetWidth(40)
				m.ShowLineNumbers = false

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line that is
					> longer than the max width
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line that is longer than the max width\nplaceholder the second line that is longer than the max width"
				m.ShowLineNumbers = false
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> placeholder the first line that is
					> longer than the max width
					> placeholder the second line that is
					> longer than the max width
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line that is longer than the max width"
				m.ShowLineNumbers = true
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line that is
					>     longer than the max width
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "placeholder the first line that is longer than the max width\nplaceholder the second line that is longer than the max width"
				m.ShowLineNumbers = true
				m.SetWidth(40)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 placeholder the first line that is
					>     longer than the max width
					>     placeholder the second line that
					>     is longer than the max width
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width at limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "123456789012345678"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width at limit plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "1234567890123456789"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 9
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers at limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "12345678901234"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder single line that is longer than max width with line numbers at limit plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "123456789012345"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     5
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width at limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "123456789012345678\n123456789012345678"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 123456789012345678
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width at limit plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "1234567890123456789\n1234567890123456789"
				m.ShowLineNumbers = false
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					> 123456789012345678
					> 9
					> 123456789012345678
					> 9
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers at limit",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "12345678901234\n12345678901234"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     12345678901234
					>
					>
					>
					>
				`),
			},
		},
		{
			name: "placeholder multiple lines that are longer than max width with line numbers at limit plus one",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "123456789012345\n123456789012345"
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 12345678901234
					>     5
					>     12345678901234
					>     5
					>
					>
				`),
			},
		},
		{
			name: "placeholder chinese character",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.Placeholder = "输入消息..."
				m.ShowLineNumbers = true
				m.SetWidth(20)

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 输入消息...
					>
					>
					>
					>
					>

				`),
			},
		},
		{
			name: "page up moves to beginning when near top",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = true
				m.SetHeight(4)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 3, 0)
				m.PageUp()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   1 Line 1
					>   2 Line 2
					>   3 Line 3
					>   4 Line 4
				`),
				cursorRow: 0,
			},
		},
		{
			name: "page up snaps to first visible line when not on it",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = true
				m.SetHeight(4)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 5, 3)
				m.PageUp()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   4 Line 4
					>   5 Line 5
					>   6 Line 6
					>   7 Line 7
				`),
				cursorRow: 3,
			},
		},
		{
			name: "page up moves up by full page when on first visible line",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.ShowLineNumbers = true
				m.SetHeight(3)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 5, 5)
				m.PageUp()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   3 Line 3
					>   4 Line 4
					>   5 Line 5
				`),
				cursorRow: 2,
			},
		},
		{
			name: "page down moves to end when near bottom",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetHeight(3)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 8, 7)
				m.PageDown()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   8 Line 8
					>   9 Line 9
					>  10 Line 10
				`),
				cursorRow: 9,
			},
		},
		{
			name: "page down snaps to last visible line when not on it",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetHeight(3)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 3, 3)
				m.PageDown()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   4 Line 4
					>   5 Line 5
					>   6 Line 6
				`),
				cursorRow: 5,
			},
		},
		{
			name: "page down moves down by full page when on last visible line",
			modelFunc: func(m textarea.Model) textarea.Model {
				m.SetHeight(3)
				m.SetWidth(20)
				m.SetValue(tenNumberedLines())

				m = scrolledTo(m, 4, 2)
				m.PageDown()

				return m
			},
			want: want{
				view: heredoc.Doc(`
					>   6 Line 6
					>   7 Line 7
					>   8 Line 8
				`),
				cursorRow: 7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newTextArea()

			if tt.modelFunc != nil {
				model = tt.modelFunc(model)
			}

			require.Equal(t, stripString(tt.want.view), stripString(model.View()))
			assert.Equal(t, tt.want.cursorRow, textarea.CursorLineNumber(&model), "cursor row")
			assert.Equal(t, tt.want.cursorCol, model.LineInfo().ColumnOffset, "cursor column")
		})
	}
}

func TestWordLeftStopsAtStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		value            string
		row, col         int
		wantRow, wantCol int
	}{
		{
			name: "empty input",
		},
		{
			name:  "buffer start before whitespace",
			value: " word",
		},
		{
			name:    "inside leading whitespace",
			value:   "   word",
			col:     2,
			wantCol: 0,
		},
		{
			name:    "multiline whitespace prefix",
			value:   "  \n  word",
			row:     1,
			col:     2,
			wantRow: 0,
			wantCol: 0,
		},
		{
			name:    "previous word",
			value:   "first second",
			col:     len("first second"),
			wantCol: len("first "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newTextArea()
			model.SetValue(tt.value)
			model = placeCursor(model, tt.row, tt.col)

			done := make(chan textarea.Model, 1)

			go func() {
				moved, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModAlt})

				done <- moved
			}()

			select {
			case got := <-done:
				assert.Equal(t, tt.wantRow, got.Line(), "cursor row")
				assert.Equal(t, tt.wantCol, got.Column(), "cursor column")
			case <-time.After(time.Second):
				t.Fatal("wordLeft did not terminate")
			}
		})
	}
}

func TestWord(t *testing.T) {
	t.Parallel()

	typed := keypress.Typed("Word1 Word2 Word3 Word4")
	navigation := []tea.KeyPressMsg{
		{Code: tea.KeyLeft, Mod: tea.ModAlt, Text: "alt+left"},
		{Code: tea.KeyLeft, Mod: tea.ModAlt, Text: "alt+left"},
		{Code: tea.KeyRight, Text: "right"},
	}
	deletion := []tea.KeyPressMsg{
		{Code: tea.KeyEnd, Text: "end"},
		{Code: tea.KeyBackspace, Mod: tea.ModAlt, Text: "alt+backspace"},
		{Code: tea.KeyBackspace, Mod: tea.ModAlt, Text: "alt+backspace"},
		{Code: tea.KeyBackspace, Text: "backspace"},
	}

	tests := []struct {
		name    string
		presses [][]tea.KeyPressMsg
		want    string
	}{
		{name: "regular input", presses: [][]tea.KeyPressMsg{typed}, want: "Word4"},
		{name: "navigate", presses: [][]tea.KeyPressMsg{typed, navigation}, want: "Word3"},
		{name: "delete", presses: [][]tea.KeyPressMsg{typed, navigation, deletion}, want: "Word2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := newTextArea()

			model.SetHeight(3)
			model.SetWidth(20)
			model.CharLimit = 500

			model, _ = model.Update(nil)

			for _, presses := range tt.presses {
				model = pressViewing(model, presses)
			}

			assert.Equal(t, tt.want, model.Word())
		})
	}
}

func newDynamicTextArea(minH, maxH int) textarea.Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.DynamicHeight = true
	ta.MinHeight = minH
	ta.MaxHeight = maxH
	ta.SetWidth(20)
	ta.Focus()
	ta, _ = ta.Update(nil)

	return ta
}

func TestDynamicHeight_DefaultUnchanged(t *testing.T) {
	t.Parallel()

	ta := newTextArea()
	ta.SetHeight(6)
	ta.SetWidth(40)

	ta = sendString(ta, "hello\nworld\n")

	assert.Equal(t, 6, ta.Height(), "expected static height 6")
}

func TestDynamicHeight_GrowsOnNewline(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	ta, _ = ta.Update(keypress.Letter('a'))
	assert.Equal(t, 1, ta.Height(), "expected height 1 after single char")

	enter := keypress.Special(tea.KeyEnter)

	ta, _ = ta.Update(enter)
	assert.Equal(t, 2, ta.Height(), "expected height 2 after first newline")

	ta, _ = ta.Update(enter)
	assert.Equal(t, 3, ta.Height(), "expected height 3 after second newline")
}

func TestDynamicHeight_GrowsOnSoftWrap(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)
	// width=20, so typing >20 chars should cause a soft wrap
	ta = sendString(ta, "abcdefghijklmnopqrstuvwxyz")

	assert.GreaterOrEqual(t, ta.Height(), 2, "expected height >= 2 after soft wrap")
}

func TestDynamicHeight_ShrinksOnLineDeletion(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	enter := keypress.Special(tea.KeyEnter)
	ta, _ = ta.Update(keypress.Letter('a'))
	ta, _ = ta.Update(enter)
	ta, _ = ta.Update(keypress.Letter('b'))
	ta, _ = ta.Update(enter)
	ta, _ = ta.Update(keypress.Letter('c'))

	require.Equal(t, 3, ta.Height(), "expected height 3 before deletion")

	// Backspace at start of line 3 merges with line 2
	ta.CursorStart()

	ta, _ = ta.Update(keypress.Special(tea.KeyBackspace))

	assert.Equal(t, 2, ta.Height(), "expected height 2 after line merge")
}

func TestDynamicHeight_RespectsMinHeight(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(5, 20)

	ta, _ = ta.Update(keypress.Letter('a'))

	assert.Equal(t, 5, ta.Height(), "expected min height 5")
}

func TestDynamicHeight_RespectsMaxHeight(t *testing.T) {
	t.Parallel()

	ta := typedLines(newDynamicTextArea(1, 5), 10)

	assert.Equal(t, 5, ta.Height(), "expected max height 5")
}

func TestDynamicHeight_GrowsOnPaste(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	paste := tea.PasteMsg{Content: "line1\nline2\nline3\nline4\nline5"}
	ta, _ = ta.Update(paste)

	assert.Equal(t, 5, ta.Height(), "expected height 5 after pasting 5 lines")
}

func TestDynamicHeight_RecalculatesOnSetWidth(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 50)
	ta.SetWidth(40)

	// Insert a line that fits in 40 cols but wraps in 10 cols
	ta.SetValue("abcdefghijklmnopqrstuvwxyz")

	require.Equal(t, 1, ta.Height(), "expected height 1 at width 40")

	ta.SetWidth(10)

	assert.GreaterOrEqual(t, ta.Height(), 3, "expected height >= 3 after narrowing to width 10")
}

func TestDynamicHeight_RecalculatesOnSetValue(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	ta.SetValue("a\nb\nc\nd\ne")

	assert.Equal(t, 5, ta.Height(), "expected height 5 after SetValue with 5 lines")
}

func TestDynamicHeight_CursorPositionAfterGrow(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	enter := keypress.Special(tea.KeyEnter)

	for i := range 5 {
		ta, _ = ta.Update(keypress.Letter(rune('a' + i)))
		ta, _ = ta.Update(enter)
	}

	ta, _ = ta.Update(keypress.Letter('f'))

	// Cursor should be on the last line (row 5, 0-indexed)
	assert.Equal(t, 5, ta.Line(), "expected cursor on row 5")
	assertCursorVisible(t, &ta, "")
}

func TestDynamicHeight_CursorPositionAfterShrink(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	enter := keypress.Special(tea.KeyEnter)

	for i := range 5 {
		ta, _ = ta.Update(keypress.Letter(rune('a' + i)))
		ta, _ = ta.Update(enter)
	}

	ta, _ = ta.Update(keypress.Letter('f'))

	require.Equal(t, 6, ta.Height(), "expected height 6 before shrink")

	// Delete lines by backspacing
	backspace := keypress.Special(tea.KeyBackspace)
	ta, _ = ta.Update(backspace) // delete 'f'
	ta, _ = ta.Update(backspace) // merge line 5 into 4
	ta, _ = ta.Update(backspace) // delete 'e'
	ta, _ = ta.Update(backspace) // merge line 4 into 3

	assertCursorVisible(t, &ta, "after shrink")
}

func TestDynamicHeight_CursorPositionAfterPaste(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 20)

	paste := tea.PasteMsg{Content: "line1\nline2\nline3\nline4\nline5"}
	ta, _ = ta.Update(paste)

	// Cursor should be at the end of the last pasted line
	assert.Equal(t, 4, ta.Line(), "expected cursor on row 4")
	assertCursorVisible(t, &ta, "after paste")
}

func TestMaxContentHeight_ScrollsBeyondMaxHeight(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 5)
	ta.MaxContentHeight = 10

	ta = typedLines(ta, 8)

	assert.Equal(t, 5, ta.Height(), "expected visible height capped at 5")
	assert.Equal(t, 9, ta.LineCount(), "expected 9 logical lines")
}

func TestMaxContentHeight_BlocksAtLimit(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.MaxContentHeight = 5
	ta.SetWidth(20)
	ta.Focus()
	ta, _ = ta.Update(nil)

	ta = typedLines(ta, 10)

	assert.LessOrEqual(t, textarea.TotalVisualLines(&ta), 5, "expected total visual lines <= 5")
}

func TestMaxContentHeight_BackwardCompat(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.MaxHeight = 10
	ta.SetWidth(20)
	ta.Focus()
	ta, _ = ta.Update(nil)

	ta = typedLines(ta, 15)

	assert.LessOrEqual(t, ta.LineCount(), 10, "expected logical line count <= 10 (legacy behavior)")
}

func TestMaxContentHeight_WithoutDynamicHeight(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.MaxContentHeight = 5
	ta.SetHeight(3)
	ta.SetWidth(20)
	ta.Focus()
	ta, _ = ta.Update(nil)

	ta = typedLines(ta, 10)

	assert.Equal(t, 3, ta.Height(), "expected fixed height 3")
	assert.LessOrEqual(t, textarea.TotalVisualLines(&ta), 5, "expected content capped at 5 visual lines")
}

func TestMaxContentHeight_CursorVisibleWhileScrolling(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 5)
	ta.MaxContentHeight = 10

	ta = typedLines(ta, 8)

	ta, _ = ta.Update(keypress.Letter('y'))

	assertCursorVisible(t, &ta, "while scrolling")
}

func TestMaxContentHeight_PasteCapped(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.MaxContentHeight = 5
	ta.SetWidth(20)
	ta.Focus()
	ta, _ = ta.Update(nil)

	paste := tea.PasteMsg{Content: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"}
	ta, _ = ta.Update(paste)

	assert.LessOrEqual(t, textarea.TotalVisualLines(&ta), 5, "expected paste capped at 5 visual lines")
}

func TestDynamicHeight_ShrinksWhenScrolledAndLinesDeleted(t *testing.T) {
	t.Parallel()

	ta := newDynamicTextArea(1, 5)
	ta.MaxContentHeight = 10

	// Type 8 lines so we exceed MaxHeight (5) and start scrolling
	ta = typedLines(ta, 7)

	ta, _ = ta.Update(keypress.Letter('x'))

	require.Equal(t, 5, ta.Height(), "expected height 5 (capped at MaxHeight)")
	require.Equal(t, 8, ta.LineCount(), "expected 8 lines")

	// Now delete lines from the bottom by selecting all on current line and backspacing
	ta = deletedLinesDownTo(ta, 4)

	// Now we have 4 lines, which is less than MaxHeight (5).
	// Height should shrink to 4.
	assert.Equal(t, 4, ta.Height(), "expected height to shrink to 4 (matching content)")
	assert.Equal(t, 0, ta.ScrollYOffset(), "expected yOffset 0 after shrinking")
}

func TestDynamicHeight_ShrinksWhenScrolledNoMaxContent(t *testing.T) {
	t.Parallel()

	// DynamicHeight with MaxHeight but no MaxContentHeight
	ta := newDynamicTextArea(1, 99)

	// Type 8 lines
	ta = typedLines(ta, 7)

	ta, _ = ta.Update(keypress.Letter('x'))

	require.Equal(t, 8, ta.Height(), "expected height 8")

	// Manually set a smaller MaxHeight to simulate scrolling scenario
	ta.MaxHeight = 5
	ta, _ = ta.Update(nil)

	// Now delete lines from the bottom
	ta = deletedLinesDownTo(ta, 3)

	assert.Equal(t, 3, ta.Height(), "expected height to shrink to 3 (matching content)")
	assert.Equal(t, 0, ta.ScrollYOffset(), "expected yOffset 0 after shrinking")
}

func newTextArea() textarea.Model {
	model := textarea.New()

	model.Prompt = "> "
	model.Placeholder = "Hello, World!"

	model.Focus()

	model, _ = model.Update(nil)

	return model
}

func sendString(m textarea.Model, str string) textarea.Model {
	for _, press := range keypress.Typed(str) {
		m, _ = m.Update(press)
	}

	return m
}

func sendStringViewing(m textarea.Model, str string) textarea.Model {
	return pressViewing(m, keypress.Typed(str))
}

func pressViewing(m textarea.Model, presses []tea.KeyPressMsg) textarea.Model {
	for _, press := range presses {
		m, _ = m.Update(press)
		m.View()
	}

	return m
}

func typedLines(ta textarea.Model, count int) textarea.Model {
	enter := keypress.Special(tea.KeyEnter)

	for range count {
		ta, _ = ta.Update(keypress.Letter('x'))
		ta, _ = ta.Update(enter)
	}

	return ta
}

func deletedLinesDownTo(ta textarea.Model, lineCount int) textarea.Model {
	backspace := keypress.Special(tea.KeyBackspace)

	for ta.LineCount() > lineCount {
		ta.CursorEnd()

		for ta.Column() > 0 {
			ta, _ = ta.Update(backspace)
		}

		ta, _ = ta.Update(backspace) // merge with previous line
	}

	return ta
}

func placeCursor(m textarea.Model, row, col int) textarea.Model {
	m.MoveToBegin()

	for range row {
		m.CursorDown()
	}

	m.SetCursorColumn(col)

	return m
}

func scrolledTo(m textarea.Model, row, yOffset int) textarea.Model {
	m.MoveToBegin()
	m.View()

	bottomRow := yOffset + m.Height() - 1
	for range bottomRow {
		m.CursorDown()
	}

	for range bottomRow - row {
		m.CursorUp()
	}

	return m
}

func tenNumberedLines() string {
	const count = 10

	lines := make([]string, count)
	for i := range count {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}

	return strings.Join(lines, "\n")
}

func requireCursorAt(t *testing.T, model textarea.Model, want textarea.Position, msg string) {
	t.Helper()

	require.Equal(t, want, textarea.Position{Row: model.Line(), Col: model.Column()}, msg)
}

func assertCursorVisible(t *testing.T, ta *textarea.Model, when string) {
	t.Helper()

	cursorLine := textarea.CursorLineNumber(ta)
	minVisible := ta.ScrollYOffset()
	maxVisible := minVisible + ta.Height() - 1

	assert.GreaterOrEqual(t, cursorLine, minVisible, "cursor line above viewport %s", when)
	assert.LessOrEqual(t, cursorLine, maxVisible, "cursor line below viewport %s", when)
}

func stripString(str string) string {
	s := ansi.Strip(str)
	ss := strings.Split(s, "\n")

	var lines []string

	for _, l := range ss {
		trim := strings.TrimRightFunc(l, unicode.IsSpace)
		if trim != "" {
			lines = append(lines, trim)
		}
	}

	return strings.Join(lines, "\n")
}

func TestSelectionCharacterForward(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world")

	// Move cursor to start.
	model.CursorStart()

	// Select "hello" by pressing shift+right 5 times.
	model = selectedRight(model, 5)

	assert.Equal(t, "hello", model.SelectedText())
	assert.True(t, model.HasSelection(), "Expected selection to be active")
}

func TestSelectionCharacterBackward(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello")

	// Cursor is at end. Select "hello" backwards.
	model = selectedLeft(model, 5)

	assert.Equal(t, "hello", model.SelectedText())
}

func TestSelectionWordForward(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world foo")

	model.CursorStart()

	// Select "hello" with alt+shift+right.
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift | tea.ModAlt})

	assert.Equal(t, "hello", model.SelectedText())
}

func TestSelectionMultiLine(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(10)

	model = sendString(model, "line one")
	model, _ = model.Update(keypress.Special(tea.KeyEnter))
	model = sendString(model, "line two")

	// Go to start.
	model.MoveToBegin()

	// Select down one line.
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown, Mod: tea.ModShift})

	assert.Equal(t, "line one\n", model.SelectedText())

	// Now select to end of line two.
	model = selectedRight(model, 8)

	assert.Equal(t, "line one\nline two", model.SelectedText())
}

func TestSelectAll(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(10)

	model = sendString(model, "hello world")

	model.SelectAll()

	assert.Equal(t, "hello world", model.SelectedText())
}

func TestSelectionClearedOnMovement(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello")
	model.CursorStart()

	// Select some text.
	model = selectedRight(model, 1)
	require.True(t, model.HasSelection(), "Expected selection to be active")

	// Move without shift, selection should clear.
	model, _ = model.Update(keypress.Special(tea.KeyRight))
	assert.False(t, model.HasSelection(), "Expected selection to be cleared after non-selection movement")
}

func TestSelectionClearedOnType(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello")
	model.CursorStart()

	// Select some text.
	model = selectedRight(model, 1)
	require.True(t, model.HasSelection(), "Expected selection to be active")

	// Type a character, selection should clear.
	model, _ = model.Update(keypress.Letter('x'))
	assert.False(t, model.HasSelection(), "Expected selection to be cleared after typing")
}

func TestClearSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello")
	model.SelectAll()

	require.True(t, model.HasSelection(), "Expected selection to be active")

	model.ClearSelection()

	assert.False(t, model.HasSelection(), "Expected selection to be cleared")
	assert.Empty(t, model.SelectedText())
}

func TestSelectionRendering(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello")

	model.CursorStart()

	model = selectedRight(model, 3)

	// Render the view and check it doesn't panic.
	assert.NotEmpty(t, model.View())
}

func TestTypingReplacesSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world")
	model.CursorStart()

	// Select "hello".
	model = selectedRight(model, 5)

	require.Equal(t, "hello", model.SelectedText())

	// Type 'X' — should replace selection.
	model, _ = model.Update(keypress.Letter('X'))

	assert.Equal(t, "X world", model.Value())
	assert.False(t, model.HasSelection(), "Expected selection to be cleared after typing")
}

func TestBackspaceDeletesSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world")
	model.CursorStart()

	// Select "hello ".
	model = selectedRight(model, 6)

	// Press backspace — should delete selection.
	model, _ = model.Update(keypress.Special(tea.KeyBackspace))

	assert.Equal(t, "world", model.Value())
	assert.False(t, model.HasSelection(), "Expected selection to be cleared")
}

func TestDeleteForwardDeletesSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world")
	model.CursorStart()

	model = selectedRight(model, 5)

	model, _ = model.Update(keypress.Special(tea.KeyDelete))

	assert.Equal(t, " world", model.Value())
}

func TestPasteReplacesSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)
	model.SetWidth(40)

	model = sendString(model, "hello world")
	model.CursorStart()

	// Select "hello".
	model = selectedRight(model, 5)

	model, _ = model.Update(tea.PasteMsg{Content: "REPLACED"})

	assert.Equal(t, "REPLACED world", model.Value())
}

func TestNewlineReplacesSelection(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(10)

	model = sendString(model, "hello world")
	model.CursorStart()

	// Select "hello ".
	model = selectedRight(model, 6)

	// Press enter — should replace selection with newline.
	model, _ = model.Update(keypress.Special(tea.KeyEnter))

	assert.Equal(t, "\nworld", model.Value())
}

func TestDeleteSelectionMultiLine(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(10)

	model = sendString(model, "line one")
	model, _ = model.Update(keypress.Special(tea.KeyEnter))
	model = sendString(model, "line two")
	model, _ = model.Update(keypress.Special(tea.KeyEnter))
	model = sendString(model, "line three")

	// Go to start and select through "line two\n".
	model.MoveToBegin()
	// Move right through "line one\n" (9 chars).
	model = selectedRight(model, 9)
	// Move right through "line two\n" (9 chars).
	model = selectedRight(model, 9)

	require.Equal(t, "line one\nline two\n", model.SelectedText())

	// Delete the selection.
	model, _ = model.Update(keypress.Special(tea.KeyBackspace))

	assert.Equal(t, "line three", model.Value())
}

func TestSelectionRangeOrdering(t *testing.T) {
	t.Parallel()

	model := newSelectingTextArea(5)

	model = sendString(model, "hello world")

	// Select backwards from end.
	model = selectedLeft(model, 5)

	// Selection should be "world" even though we selected backwards.
	assert.Equal(t, "world", model.SelectedText())
}

func TestKeyMapRebindingAfterNew(t *testing.T) {
	t.Parallel()

	model := newTextArea()
	model.KeyMap.LineStart = key.NewBinding(key.WithKeys("ctrl+z"))

	model = sendString(model, "hello")
	model, _ = model.Update(keypress.Ctrl('z'))

	assert.Equal(t, 0, model.Column())
}

func newSelectingTextArea(height int) textarea.Model {
	model := newTextArea()
	model.Prompt = ""
	model.ShowLineNumbers = false
	model.SetWidth(20)
	model.SetHeight(height)

	return model
}

func selectedRight(m textarea.Model, count int) textarea.Model {
	return pressed(m, tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, count)
}

func selectedLeft(m textarea.Model, count int) textarea.Model {
	return pressed(m, tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModShift}, count)
}

func pressed(m textarea.Model, press tea.KeyPressMsg, count int) textarea.Model {
	for range count {
		m, _ = m.Update(press)
	}

	return m
}

func TestModel_SetValue_maxLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxLines int
		lines    int
		want     int
	}{
		{
			name:     "cuts upstream's default at 10,000 lines",
			maxLines: textarea.New().MaxLines,
			lines:    10_001,
			want:     10_000,
		},
		{name: "cuts at the set number of lines", maxLines: 3, lines: 4, want: 3},
		{name: "keeps every line with no limit set", maxLines: 0, lines: 20_000, want: 20_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			area := textarea.New()
			area.MaxLines = tt.maxLines
			area.SetValue(strings.Repeat("line\n", tt.lines-1) + "line")

			assert.Equal(t, tt.want, area.LineCount())
		})
	}
}

func TestModel_Value_tabs(t *testing.T) {
	t.Parallel()

	t.Run("keeps loaded tabs", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "\tif x {\n\t\treturn\n\t}")

		assert.Equal(t, "\tif x {\n\t\treturn\n\t}", area.Value())
	})

	t.Run("keeps inserted tabs", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "")
		area.InsertString("a\tb")

		assert.Equal(t, "a\tb", area.Value())
	})

	t.Run("keeps pasted tabs", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "")
		area, _ = area.Update(tea.PasteMsg{Content: "a\tb\n\tc"})

		assert.Equal(t, "a\tb\n\tc", area.Value())
	})
}

func TestModel_View_tabs(t *testing.T) {
	t.Parallel()

	t.Run("runs a tab to the next tab stop", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "a\tb\nabc\td\n\t\te")

		assert.Equal(t, []string{"a   b", "abc d", "        e"}, viewRows(&area)[:3])
	})

	t.Run("wraps a line with tabs at the tab-aware width", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "\t\t\t\tword\tnext")

		assert.Equal(t, []string{"", "word    next"}, viewRows(&area)[:2])
		assert.Equal(t, "\t\t\t\tword\tnext", area.Value())
	})

	t.Run("keeps a word on its row when only the tab after it overflows", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "top\nabcdefgh\tx")
		area.SetWidth(10)

		assert.Equal(t, []string{"top", "abcdefgh", "x"}, viewRows(&area)[:3])
	})

	t.Run("counts the tab stops on a wrapped row from the start of the line", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abcdef ghij\tk")
		area.SetWidth(10)

		assert.Equal(t, []string{"abcdef", "ghij k"}, viewRows(&area)[:2])
	})

	t.Run("draws a wrapped row's tab on the cells the Snippet pane draws it on", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, wrappedTabLine)
		area.SetWidth(10)

		want := []string{"abcdefghi", "jk    l"}
		assert.Equal(t, want, viewRows(&area)[:2])
		assert.Equal(t, want, wrappedSnippetPaneRows(t, wrappedTabLine, 10)[:2])
	})

	t.Run("covers the whole tab with the reverse-video cursor block", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "a\tx")
		area.MoveToBegin()
		area, _ = area.Update(keypress.Special(tea.KeyRight))

		firstRow, _, _ := strings.Cut(area.View(), "\n")

		assert.Contains(t, firstRow, "\x1b[7m   \x1b[m")
	})

	t.Run("covers the whole tab on a wrapped row with the cursor block", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, wrappedTabLine)
		area.SetWidth(10)
		area.SetCursorColumn(12)

		secondRow := strings.Split(area.View(), "\n")[1]

		assert.Contains(t, secondRow, "\x1b[7m    \x1b[m")
	})
}

func wrappedSnippetPaneRows(t *testing.T, content string, codeWidth int) []string {
	t.Helper()

	const (
		gutter      = "   1 │ "
		headerLines = 4
	)

	pane := snippetpane.New(testsettings.Default(t).Keys, look.NewStyles(look.SchemeDark), time.UTC)
	pane, _, _ = pane.Update(look.Resized{Box: look.Size{Width: ansi.StringWidth(gutter) + codeWidth, Height: 10}})
	pane = pane.Showing(testkit.Snippet(t, testkit.SnippetSpec{Fragments: []testkit.FragmentSpec{{Content: content}}}))
	pane, _, _ = pane.Update(keypress.Letter('w'))

	rows := strings.Split(ansi.Strip(pane.View()), "\n")[headerLines:]
	for index, row := range rows {
		_, code, _ := strings.Cut(row, "│ ")
		rows[index] = strings.TrimRight(code, " ")
	}

	return rows
}

func TestModel_Update_tabs(t *testing.T) {
	t.Parallel()

	t.Run("left from after a tab lands before it in one press", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "a\tb")
		area.SetCursorColumn(2)

		area, _ = area.Update(keypress.Special(tea.KeyLeft))

		assert.Equal(t, 1, area.Column())
	})

	t.Run("right from before a tab lands after it in one press", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "a\tb")
		area.SetCursorColumn(1)

		area, _ = area.Update(keypress.Special(tea.KeyRight))

		assert.Equal(t, 2, area.Column())
	})

	t.Run("down lands on the tab covering the cell column and keeps it", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abcdef\n\tx\nabcdef")
		area.MoveToBegin()
		area.SetCursorColumn(2)

		area, _ = area.Update(keypress.Special(tea.KeyDown))
		onTab := area.Column()
		area, _ = area.Update(keypress.Special(tea.KeyDown))

		assert.Equal(t, 0, onTab)
		assert.Equal(t, 2, area.Column())
	})

	t.Run("up from after a tab keeps the cell column", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, "abcdef\n\tx")
		area.SetCursorColumn(1)

		area, _ = area.Update(keypress.Special(tea.KeyUp))

		assert.Equal(t, 4, area.Column())
	})

	t.Run("down from after a tab on a wrapped row keeps the cell column", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, wrappedTabLine+"\n0123456789")
		area.SetWidth(10)
		area.MoveToBegin()
		area.SetCursorColumn(13)

		area, _ = area.Update(keypress.Special(tea.KeyDown))

		assert.Equal(t, 6, area.Column())
	})

	t.Run("up onto a wrapped row lands on the tab covering the cell column", func(t *testing.T) {
		t.Parallel()

		area := newPlainTextarea(t, wrappedTabLine+"\n0123456789")
		area.SetWidth(10)
		area.SetCursorColumn(5)

		area, _ = area.Update(keypress.Special(tea.KeyUp))

		assert.Equal(t, 12, area.Column())
	})
}
