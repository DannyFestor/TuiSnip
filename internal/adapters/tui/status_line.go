package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

const (
	hintSeparator = " · "
	tooSmallHint  = "Terminal too small for all four Panes (80×24)"
)

func statusLine(styles look.Styles, message, hint string, width int) string {
	hintWidth := ansi.StringWidth(hint)
	left := ansi.Truncate(" "+message, max(0, width-hintWidth-1), look.Ellipsis)
	gap := strings.Repeat(" ", max(1, width-ansi.StringWidth(left)-hintWidth))

	return look.FitWidth(left+gap+styles.Dim.Render(hint), width)
}

func hintFor(hints []key.Binding, width int) string {
	entries := make([]string, 0, len(hints))

	for _, hint := range hints {
		if hint.Enabled() {
			entries = append(entries, hint.Help().Key+" "+hint.Help().Desc)
		}
	}

	for len(entries) > 0 && ansi.StringWidth(strings.Join(entries, hintSeparator)) > width {
		entries = entries[:len(entries)-1]
	}

	return strings.Join(entries, hintSeparator)
}
