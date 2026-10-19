package tabstop

import (
	"strings"

	"github.com/rivo/uniseg"
)

const lineBreak = "\n"

func Expand(text string) string {
	lines := strings.Split(text, lineBreak)
	for index, line := range lines {
		lines[index] = ExpandFrom([]rune(line), 0)
	}

	return strings.Join(lines, lineBreak)
}

func ExpandFrom(runes []rune, column int) string {
	var expanded strings.Builder

	chunkStart := 0

	for index, r := range runes {
		if r != tab {
			continue
		}

		chunk := string(runes[chunkStart:index])
		expanded.WriteString(chunk)
		column += uniseg.StringWidth(chunk)
		cells := cellsOfTabAt(column)
		expanded.WriteString(strings.Repeat(" ", cells))
		column += cells
		chunkStart = index + 1
	}

	expanded.WriteString(string(runes[chunkStart:]))

	return expanded.String()
}
