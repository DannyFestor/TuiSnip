package confirm

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
	title           = "Unsaved changes"
	rows            = 1
	titlePadding    = 2
	answerSeparator = "/"
)

type Confirmation struct {
	keys     binding.Set
	styles   look.Styles
	question string
	onYes    outcome.Outcome
}

func New(keys binding.Keys, styles look.Styles, question string, onYes outcome.Outcome) Confirmation {
	return Confirmation{keys: keys.For(binding.ScopeConfirm), styles: styles, question: question, onYes: onYes}
}

func (c Confirmation) Update(msg tea.Msg) outcome.Step {
	pressed, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return outcome.Stay(c)
	}

	switch {
	case c.keys.Matches(pressed, binding.Yes):
		return outcome.Close().Passing(c.onYes)
	case c.keys.Matches(pressed, binding.No):
		return outcome.Close()
	}

	return outcome.Stay(c)
}

func (c Confirmation) View() string {
	prompt := c.prompt()
	width := max(ansi.StringWidth(prompt), ansi.StringWidth(title)+titlePadding)
	outer := look.Size{Width: width + look.BorderWidth, Height: rows + look.BorderWidth}

	return look.Frame(c.styles.Focused, title, prompt, outer)
}

func (c Confirmation) ShortHelp() []key.Binding {
	return c.keys.ShortHelp()
}

func (c Confirmation) FullHelp() [][]key.Binding {
	return c.keys.FullHelp()
}

func (c Confirmation) prompt() string {
	answers := binding.BoundOnly(c.keys.FirstKey(binding.Yes), markedAsDefault(c.keys.FirstKey(binding.No)))

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
