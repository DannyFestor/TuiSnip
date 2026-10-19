package textarea_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestModel_KeepsTabs(t *testing.T) {
	t.Parallel()

	t.Run("keeps loaded tabs", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "\tif x {\n\t\treturn\n\t}")

		assert.Equal(t, "\tif x {\n\t\treturn\n\t}", area.Value())
	})

	t.Run("keeps inserted tabs", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "")
		area.InsertString("a\tb")

		assert.Equal(t, "a\tb", area.Value())
	})

	t.Run("keeps pasted tabs", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "")
		area, _ = area.Update(tea.PasteMsg{Content: "a\tb\n\tc"})

		assert.Equal(t, "a\tb\n\tc", area.Value())
	})
}

func TestModel_ViewTabs(t *testing.T) {
	t.Parallel()

	t.Run("runs a tab to the next tab stop", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tb\nabc\td\n\t\te")

		assert.Equal(t, []string{"a   b", "abc d", "        e"}, viewRows(&area)[:3])
	})

	t.Run("wraps a line with tabs at the tab-aware width", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "\t\t\t\tword\tnext")

		assert.Equal(t, []string{"", "word    next"}, viewRows(&area)[:2])
		assert.Equal(t, "\t\t\t\tword\tnext", area.Value())
	})
}

func TestModel_ViewCursorOnTab(t *testing.T) {
	t.Parallel()

	area := newSelectionTextarea(t, "a\tx")
	area.MoveToBegin()
	area, _ = area.Update(keypress.Special(tea.KeyRight))

	firstRow, _, _ := strings.Cut(area.View(), "\n")

	assert.Contains(t, firstRow, "\x1b[7m   \x1b[m")
}

func TestModel_UpdateMovesOverTabs(t *testing.T) {
	t.Parallel()

	t.Run("left from after a tab lands before it in one press", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tb")
		area.SetCursorColumn(2)

		area, _ = area.Update(keypress.Special(tea.KeyLeft))

		assert.Equal(t, 1, area.Column())
	})

	t.Run("right from before a tab lands after it in one press", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tb")
		area.SetCursorColumn(1)

		area, _ = area.Update(keypress.Special(tea.KeyRight))

		assert.Equal(t, 2, area.Column())
	})

	t.Run("down lands on the tab covering the cell column and keeps it", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "abcdef\n\tx\nabcdef")
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

		area := newSelectionTextarea(t, "abcdef\n\tx")
		area.SetCursorColumn(1)

		area, _ = area.Update(keypress.Special(tea.KeyUp))

		assert.Equal(t, 4, area.Column())
	})
}

func TestModel_SelectsAcrossTabs(t *testing.T) {
	t.Parallel()

	t.Run("gives the tab back from a keyboard selection and draws it at the tab stop", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tb")
		area.CursorStart()

		for range 3 {
			area, _ = area.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift})
		}

		assert.Equal(t, "a\tb", area.SelectedText())
		assert.Equal(t, "a   b", viewRows(&area)[0])
	})

	t.Run("gives the tab back from a mouse drag ending on the cell after it", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tbc")
		area.BeginSelection(0, 0)
		area.ExtendSelection(5, 0)
		area.EndSelection()

		assert.Equal(t, "a\tb", area.SelectedText())
	})

	t.Run("maps every cell a tab covers onto the tab", func(t *testing.T) {
		t.Parallel()

		area := newSelectionTextarea(t, "a\tb")

		for cell := 1; cell < 4; cell++ {
			assert.Equal(t, textarea.Position{Row: 0, Col: 1}, area.PositionAt(cell, 0), "cell %d", cell)
		}

		assert.Equal(t, textarea.Position{Row: 0, Col: 2}, area.PositionAt(4, 0))
	})
}

func viewRows(area *textarea.Model) []string {
	rows := strings.Split(ansi.Strip(area.View()), "\n")
	for index, row := range rows {
		rows[index] = strings.TrimRight(row, " ")
	}

	return rows
}
