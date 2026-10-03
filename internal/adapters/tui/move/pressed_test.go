package move_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPressed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pressed tea.KeyPressMsg
		want    move.Direction
		wantOK  bool
	}{
		{name: "down", pressed: letter('j'), want: move.Down, wantOK: true},
		{name: "up", pressed: letter('k'), want: move.Up, wantOK: true},
		{name: "top", pressed: letter('g'), want: move.Top, wantOK: true},
		{name: "bottom", pressed: letter('G'), want: move.Bottom, wantOK: true},
		{name: "page down", pressed: tea.KeyPressMsg{Code: tea.KeyPgDown}, want: move.PageDown, wantOK: true},
		{name: "page up", pressed: tea.KeyPressMsg{Code: tea.KeyPgUp}, want: move.PageUp, wantOK: true},
		{name: "no movement for other keys", pressed: letter('x'), want: move.None, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			global := testsettings.Default(t).Keys.For(binding.ScopeGlobal)

			got, ok := move.Pressed(global, tt.pressed)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}
