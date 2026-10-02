package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
)

const (
	hintSeparator = " · "
	tooSmallHint  = "Terminal too small for all four Panes (80×24)"
)

func statusLine(styles styleSet, message, hint string, width int) string {
	hintWidth := ansi.StringWidth(hint)
	left := ansi.Truncate(" "+message, max(0, width-hintWidth-1), ellipsis)
	gap := strings.Repeat(" ", max(1, width-ansi.StringWidth(left)-hintWidth))

	return fitWidth(left+gap+styles.dim.Render(hint), width)
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
