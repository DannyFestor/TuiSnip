package look

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	BorderWidth  = 2
	Ellipsis     = "…"
	titlePadding = 2
)

func Frame(style FrameStyle, title, body string, outer Size) string {
	inner := outer.Inner()
	lines := make([]string, 0, outer.Height)
	lines = append(lines, topBorder(style, title, inner.Width))

	for _, line := range fitLines(body, inner) {
		lines = append(lines, style.Border.Render("│")+line+style.Border.Render("│"))
	}

	lines = append(lines, style.Border.Render("╰"+strings.Repeat("─", inner.Width)+"╯"))

	return strings.Join(lines, "\n")
}

func FitWidth(line string, width int) string {
	truncated := ansi.Truncate(line, width, Ellipsis)

	return truncated + strings.Repeat(" ", max(0, width-ansi.StringWidth(truncated)))
}

func Row(text, meta string, width int) string {
	metaWidth := ansi.StringWidth(meta)
	textWidth := max(0, width-metaWidth-1)
	truncated := ansi.Truncate(text, textWidth, Ellipsis)
	gap := max(1, width-ansi.StringWidth(truncated)-metaWidth)

	return FitWidth(truncated+strings.Repeat(" ", gap)+meta, width)
}

func topBorder(style FrameStyle, title string, innerWidth int) string {
	label := titleLabel(title, innerWidth)
	fill := max(0, innerWidth-ansi.StringWidth(label))

	return style.Border.Render("╭") + style.Title.Render(label) + style.Border.Render(strings.Repeat("─", fill)+"╮")
}

func titleLabel(title string, innerWidth int) string {
	if title == "" {
		return ""
	}

	return " " + ansi.Truncate(title, max(0, innerWidth-titlePadding), Ellipsis) + " "
}

func fitLines(body string, inner Size) []string {
	lines := strings.Split(body, "\n")
	fitted := make([]string, inner.Height)

	for index := range fitted {
		line := ""
		if index < len(lines) {
			line = lines[index]
		}

		fitted[index] = FitWidth(line, inner.Width)
	}

	return fitted
}
