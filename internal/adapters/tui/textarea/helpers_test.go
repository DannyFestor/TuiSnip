package textarea_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
)

func newPlainTextarea(t *testing.T, value string) textarea.Model {
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

func viewRows(area *textarea.Model) []string {
	rows := strings.Split(ansi.Strip(area.View()), "\n")
	for index, row := range rows {
		rows[index] = strings.TrimRight(row, " ")
	}

	return rows
}
