package screencell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func Find(tb testing.TB, screen, text string) pointer.Point {
	tb.Helper()

	for row, line := range strings.Split(ansi.Strip(screen), "\n") {
		if before, _, found := strings.Cut(line, text); found {
			return pointer.Point{X: ansi.StringWidth(before), Y: row}
		}
	}

	tb.Fatalf("%q is not on the screen", text)

	return pointer.Point{X: 0, Y: 0}
}
