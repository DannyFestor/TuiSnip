package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
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
	onYes    outcome.Outcome
}

func newConfirmation(keys confirmBindings, styles styleSet, question string, onYes outcome.Outcome) confirmation {
	return confirmation{keys: keys, styles: styles, question: question, onYes: onYes}
}

func (c confirmation) Update(msg tea.Msg) step {
	pressed, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return stay(c)
	}

	switch {
	case key.Matches(pressed, c.keys.yes):
		return closing().Passing(c.onYes)
	case key.Matches(pressed, c.keys.no):
		return closing()
	}

	return stay(c)
}

func (c confirmation) View() string {
	width := max(ansi.StringWidth(c.question), ansi.StringWidth(confirmationTitle)+confirmationPadding)
	outer := size{width: width + borderWidth, height: confirmationRows + borderWidth}

	return frame(c.styles.focused, confirmationTitle, c.question, outer)
}

func (c confirmation) Hints() []key.Binding {
	return c.keys.hints()
}
