package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	borderWidth  = 2
	ellipsis     = "…"
	titlePadding = 2
)

func frame(look paneLook, title, body string, outer size) string {
	inner := innerSize(outer)
	lines := make([]string, 0, outer.height)
	lines = append(lines, topBorder(look, title, inner.width))

	for _, line := range fitLines(body, inner) {
		lines = append(lines, look.border.Render("│")+line+look.border.Render("│"))
	}

	lines = append(lines, look.border.Render("╰"+strings.Repeat("─", inner.width)+"╯"))

	return strings.Join(lines, "\n")
}

func innerSize(outer size) size {
	return size{width: max(0, outer.width-borderWidth), height: max(0, outer.height-borderWidth)}
}

func topBorder(look paneLook, title string, innerWidth int) string {
	label := " " + ansi.Truncate(title, max(0, innerWidth-titlePadding), ellipsis) + " "
	fill := max(0, innerWidth-ansi.StringWidth(label))

	return look.border.Render("╭") + look.title.Render(label) + look.border.Render(strings.Repeat("─", fill)+"╮")
}

func fitLines(body string, inner size) []string {
	lines := strings.Split(body, "\n")
	fitted := make([]string, inner.height)

	for index := range fitted {
		line := ""
		if index < len(lines) {
			line = lines[index]
		}

		fitted[index] = fitWidth(line, inner.width)
	}

	return fitted
}

func fitWidth(line string, width int) string {
	truncated := ansi.Truncate(line, width, ellipsis)

	return truncated + strings.Repeat(" ", max(0, width-ansi.StringWidth(truncated)))
}

func row(text, meta string, width int) string {
	metaWidth := ansi.StringWidth(meta)
	textWidth := max(0, width-metaWidth-1)
	truncated := ansi.Truncate(text, textWidth, ellipsis)
	gap := max(1, width-ansi.StringWidth(truncated)-metaWidth)

	return fitWidth(truncated+strings.Repeat(" ", gap)+meta, width)
}
