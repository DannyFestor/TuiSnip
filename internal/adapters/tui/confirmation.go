package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/overlay"
)

const (
	confirmationTitle   = "Unsaved changes"
	confirmationRows    = 1
	confirmationPadding = 2
	discardQuestion     = "Discard the unsaved changes? y/N"
	quitQuestion        = "Quit and discard the unsaved changes? y/N"
)

type confirmation struct {
	keys     confirmBindings
	styles   styleSet
	question string
	onYes    overlay.Outcome
}

func newConfirmation(keys confirmBindings, styles styleSet, question string, onYes overlay.Outcome) confirmation {
	return confirmation{keys: keys, styles: styles, question: question, onYes: onYes}
}

func (c confirmation) Update(msg tea.Msg) overlay.Step {
	pressed, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return overlay.Stay(c)
	}

	switch {
	case key.Matches(pressed, c.keys.yes):
		return overlay.Close().Passing(c.onYes)
	case key.Matches(pressed, c.keys.no):
		return overlay.Close()
	}

	return overlay.Stay(c)
}

func (c confirmation) View() string {
	width := max(ansi.StringWidth(c.question), ansi.StringWidth(confirmationTitle)+confirmationPadding)
	outer := size{width: width + borderWidth, height: confirmationRows + borderWidth}

	return frame(c.styles.focused, confirmationTitle, c.question, outer)
}

func (c confirmation) Hints() []key.Binding {
	return c.keys.hints()
}
