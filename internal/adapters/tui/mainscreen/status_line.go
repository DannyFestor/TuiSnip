package mainscreen

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

const (
	hintSeparator = " · "
	messageIndent = " "
	minimumGap    = 1
	tooSmallHint  = "Terminal too small for all four Panes (80×24)"
)

func statusLine(styles look.Styles, message, hint string, width int) string {
	hintWidth := ansi.StringWidth(hint)
	left := ansi.Truncate(messageIndent+message, max(0, width-hintWidth-minimumGap), look.Ellipsis)
	gap := strings.Repeat(" ", max(minimumGap, width-ansi.StringWidth(left)-hintWidth))

	return look.FitWidth(left+gap+styles.Dim.Render(hint), width)
}

func hintRoom(message string, width int) int {
	return width - ansi.StringWidth(messageIndent+message) - minimumGap
}

func hintFor(hints []key.Binding, width int) string {
	shown := enabledOnly(hints)

	for ansi.StringWidth(hintText(shown)) > width {
		dropping := lastDroppable(shown)
		if dropping < 0 {
			break
		}

		shown = slices.Delete(shown, dropping, dropping+1)
	}

	return hintText(shown)
}

func enabledOnly(hints []key.Binding) []key.Binding {
	return slices.DeleteFunc(slices.Clone(hints), func(hint key.Binding) bool { return !hint.Enabled() })
}

func lastDroppable(hints []key.Binding) int {
	for index, hint := range slices.Backward(hints) {
		if !binding.AlwaysShown(hint) {
			return index
		}
	}

	return -1
}

func hintText(hints []key.Binding) string {
	entries := make([]string, 0, len(hints))
	for _, hint := range hints {
		entries = append(entries, hint.Help().Key+" "+hint.Help().Desc)
	}

	return strings.Join(entries, hintSeparator)
}
