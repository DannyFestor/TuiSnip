package snippetpane_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
)

func TestPane_UpdateWheel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []int
		want  string
	}{
		{name: "turning down scrolls down", lines: []int{3}, want: "   4 │ line 4"},
		{name: "turning up scrolls back up", lines: []int{3, -3}, want: "   1 │ line 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pane := showing(t, longSnippet(t))
			for _, lines := range tt.lines {
				pane, _, _ = pane.Update(pointer.Wheeled{At: pointer.Point{X: 0, Y: 0}, Lines: lines})
			}

			assert.Equal(t, tt.want, trimmed(code(pane))[0])
		})
	}
}
