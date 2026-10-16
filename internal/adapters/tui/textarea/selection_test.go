package textarea_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func newSelectionTextarea(t *testing.T, value string) textarea.Model {
	t.Helper()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(20)
	ta.SetHeight(6)
	ta.Focus()
	ta.SetValue(value)

	return ta
}

func TestPositionAt(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello\nworld")

	tests := []struct {
		name string
		x, y int
		want textarea.Position
	}{
		{"origin", 0, 0, textarea.Position{Row: 0, Col: 0}},
		{"mid first line", 3, 0, textarea.Position{Row: 0, Col: 3}},
		{"start second line", 0, 1, textarea.Position{Row: 1, Col: 0}},
		{"mid second line", 2, 1, textarea.Position{Row: 1, Col: 2}},
		{"past end of line clamps to line end", 40, 0, textarea.Position{Row: 0, Col: 5}},
		{"negative y clamps to buffer start", 0, -3, textarea.Position{Row: 0, Col: 0}},
		{"below last line clamps to buffer end", 0, 50, textarea.Position{Row: 1, Col: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ta.PositionAt(tt.x, tt.y), "PositionAt(%d, %d)", tt.x, tt.y)
		})
	}
}

func TestPositionAtAccountsForGutter(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = "> "
	ta.ShowLineNumbers = false
	ta.SetWidth(20)
	ta.SetHeight(3)
	ta.Focus()
	ta.SetValue("hello")

	assert.Equal(t, textarea.Position{Row: 0, Col: 0}, ta.PositionAt(2, 0), "first text cell")
	assert.Equal(t, textarea.Position{Row: 0, Col: 3}, ta.PositionAt(5, 0), "fourth text cell")
	assert.Equal(t, textarea.Position{Row: 0, Col: 0}, ta.PositionAt(0, 0), "click in gutter")
}

func TestSelectedTextSingleLine(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()

	require.True(t, ta.HasSelection(), "expected an active selection")
	assert.Equal(t, "hello", ta.SelectedText())
}

func TestSelectedTextAcrossLines(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "one\ntwo\nthree")
	ta.BeginSelection(2, 0)
	ta.ExtendSelection(1, 2)
	ta.EndSelection()

	assert.Equal(t, "e\ntwo\nt", ta.SelectedText())
}

func TestSelectedTextBackwardDrag(t *testing.T) {
	t.Parallel()

	forward := newSelectionTextarea(t, "one\ntwo\nthree")
	forward.BeginSelection(2, 0)
	forward.ExtendSelection(1, 2)
	forward.EndSelection()

	backward := newSelectionTextarea(t, "one\ntwo\nthree")
	backward.BeginSelection(1, 2)
	backward.ExtendSelection(2, 0)
	backward.EndSelection()

	assert.Equal(t, forward.SelectedText(), backward.SelectedText(), "backward drag must match the forward one")
	assert.Equal(t, "e\ntwo\nt", backward.SelectedText())
}

func TestClickWithoutDragMakesNoSelection(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(3, 0)
	ta.EndSelection()

	assert.False(t, ta.HasSelection(), "a click without a drag must not leave a selection")
	assert.Empty(t, ta.SelectedText())
}

func TestExtendSelectionRequiresBegin(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.ExtendSelection(5, 0)

	assert.False(t, ta.HasSelection(), "ExtendSelection without BeginSelection must not select")
}

func TestClearSelectionAPI(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()

	require.True(t, ta.HasSelection(), "expected a selection to clear")

	ta.ClearSelection()

	assert.False(t, ta.HasSelection(), "HasSelection() = true after ClearSelection()")
	assert.Empty(t, ta.SelectedText())
}

func TestSelectAllAPI(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "one\ntwo")
	ta.SelectAll()

	assert.Equal(t, "one\ntwo", ta.SelectedText())
}

func TestKeyPressClearsSelection(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()

	require.True(t, ta.HasSelection(), "expected a selection before typing")

	ta, _ = ta.Update(keypress.Letter('x'))

	assert.False(t, ta.HasSelection(), "selection must be cleared by keyboard input")
}

func TestSelectionRendersWithSelectionStyle(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")

	plain := ta.View()

	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()
	selected := ta.View()

	require.NotEqual(t, plain, selected, "view must change when a selection is active")
	assert.Equal(t, strings.TrimSpace(ansi.Strip(plain)), strings.TrimSpace(ansi.Strip(selected)))
}

func TestSelectionSurvivesSoftWrap(t *testing.T) {
	t.Parallel()

	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(10)
	ta.SetHeight(6)
	ta.Focus()
	ta.SetValue("aaaaabbbbbccccc")

	pos := ta.PositionAt(0, 1)
	require.Equal(t, 0, pos.Row, "soft wrap must stay on logical row 0")
	assert.Equal(t, 10, pos.Col, "second visual row starts at col 10")

	ta.BeginSelection(0, 1)
	ta.ExtendSelection(5, 1)
	ta.EndSelection()

	assert.Equal(t, "ccccc", ta.SelectedText())
}

func TestSelectionEmptyBuffer(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 3)
	ta.EndSelection()
	ta.SelectAll()

	assert.False(t, ta.HasSelection(), "empty buffer must not report a selection")
	assert.Empty(t, ta.SelectedText())

	_ = ta.View()
}

func TestSelectionRange(t *testing.T) {
	t.Parallel()

	ta := newSelectionTextarea(t, "hello world")
	ta.BeginSelection(0, 0)
	ta.ExtendSelection(5, 0)
	ta.EndSelection()

	selected, ok := ta.Selection()
	require.True(t, ok, "expected a selection")
	assert.Equal(t, textarea.Position{Row: 0, Col: 0}, selected.Start)
	assert.Equal(t, textarea.Position{Row: 0, Col: 5}, selected.End)
}
