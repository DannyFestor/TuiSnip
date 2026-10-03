package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

const (
	confirmationTitle   = "Unsaved changes"
	confirmationRows    = 1
	confirmationPadding = 2
	discardQuestion     = "Discard the unsaved changes?"
	quitQuestion        = "Quit and discard the unsaved changes?"
	answerSeparator     = "/"
)

type confirmation struct {
	keys     binding.Set
	styles   look.Styles
	question string
	onYes    outcome.Outcome
}

func newConfirmation(keys binding.Set, styles look.Styles, question string, onYes outcome.Outcome) confirmation {
	return confirmation{keys: keys, styles: styles, question: question, onYes: onYes}
}

func (c confirmation) Update(msg tea.Msg) step {
	pressed, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return stay(c)
	}

	switch {
	case c.keys.Matches(pressed, binding.Yes):
		return closing().Passing(c.onYes)
	case c.keys.Matches(pressed, binding.No):
		return closing()
	}

	return stay(c)
}

func (c confirmation) View() string {
	prompt := c.prompt()
	width := max(ansi.StringWidth(prompt), ansi.StringWidth(confirmationTitle)+confirmationPadding)
	outer := look.Size{Width: width + look.BorderWidth, Height: confirmationRows + look.BorderWidth}

	return look.Frame(c.styles.Focused, confirmationTitle, prompt, outer)
}

func (c confirmation) ShortHelp() []key.Binding {
	return c.keys.ShortHelp()
}

func (c confirmation) FullHelp() [][]key.Binding {
	return c.keys.FullHelp()
}

func (c confirmation) prompt() string {
	answers := boundOnly(c.keys.FirstKey(binding.Yes), markedAsDefault(c.keys.FirstKey(binding.No)))

	if len(answers) == 0 {
		return c.question
	}

	return c.question + " " + strings.Join(answers, answerSeparator)
}

func markedAsDefault(answer string) string {
	character, size := utf8.DecodeRuneInString(answer)
	if size == 0 || size != len(answer) || !unicode.IsPrint(character) {
		return answer
	}

	return strings.ToUpper(answer)
}
