package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	confirmationTitle   = "Unsaved changes"
	confirmationRows    = 1
	confirmationPadding = 2
	discardQuestion     = "Discard the unsaved changes? y/N"
	quitQuestion        = "Quit and discard the unsaved changes? y/N"
)

type confirmAnswer int

const (
	answerPending confirmAnswer = iota
	answerYes
	answerNo
)

type confirmation struct {
	open     bool
	keys     confirmBindings
	question string
	onYes    func(Model) (Model, tea.Cmd)
}

func newConfirmation(keys confirmBindings, question string, onYes func(Model) (Model, tea.Cmd)) confirmation {
	return confirmation{open: true, keys: keys, question: question, onYes: onYes}
}

func noConfirmation() confirmation {
	return confirmation{}
}

func (c confirmation) answer(msg tea.KeyPressMsg) confirmAnswer {
	switch {
	case key.Matches(msg, c.keys.yes):
		return answerYes
	case key.Matches(msg, c.keys.no):
		return answerNo
	}

	return answerPending
}

func (c confirmation) view(styles styleSet) string {
	width := max(ansi.StringWidth(c.question), ansi.StringWidth(confirmationTitle)+confirmationPadding)
	outer := size{width: width + borderWidth, height: confirmationRows + borderWidth}

	return frame(styles.focused, confirmationTitle, c.question, outer)
}

func (c confirmation) hints() []key.Binding {
	return c.keys.hints()
}
