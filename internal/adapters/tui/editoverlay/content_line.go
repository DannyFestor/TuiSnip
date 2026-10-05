package editoverlay

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

const indentLevel = "    "

func indentedLine(area textarea.Model) textarea.Model {
	column := area.Column()

	area.ClearSelection()
	area.CursorStart()
	area.InsertString(indentLevel)
	area.SetCursorColumn(column + len(indentLevel))

	return area
}

func dedentedLine(area textarea.Model) (textarea.Model, tea.Cmd) {
	removed := removableIndent(cursorLine(area))
	if removed == 0 {
		return area, nil
	}

	column := area.Column()

	area.ClearSelection()
	area.SetCursorColumn(removed)

	area, cmd := area.Update(deleteBeforeCursorPress())
	area.SetCursorColumn(max(0, column-removed))

	return area, cmd
}

func cursorLine(area textarea.Model) string {
	return strings.Split(area.Value(), "\n")[area.Line()]
}

func removableIndent(line string) int {
	indent := len(line) - len(strings.TrimLeft(line, " "))

	return min(indent, len(indentLevel))
}

// The textarea has no method that deletes text, so dedent sends it the press
// of its default DeleteBeforeCursor key.
func deleteBeforeCursorPress() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: 'u', Text: "", Mod: tea.ModCtrl, ShiftedCode: 0, BaseCode: 0, IsRepeat: false}
}
