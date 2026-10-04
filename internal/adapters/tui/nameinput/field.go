package nameinput

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
)

type Field struct {
	keys     binding.Set
	input    textinput.Model
	validate func(raw string) error
}

func New(keys binding.Keys, validate func(raw string) error, initial string) (Field, tea.Cmd) {
	line := input.NewLine("")
	line.SetValue(initial)
	cmd := line.Focus()

	return Field{keys: keys.For(binding.ScopeNameInput), input: line, validate: validate}, cmd
}

func (f Field) Update(msg tea.Msg) (Field, Result, tea.Cmd) {
	if pressed, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case f.keys.Matches(pressed, binding.Accept):
			return f, f.committed(), nil
		case f.keys.Matches(pressed, binding.Cancel):
			return f, cancelled(), nil
		}
	}

	next := f

	var cmd tea.Cmd

	next.input, cmd = f.input.Update(msg)

	return next, typing(), cmd
}

func (f Field) View() string {
	return f.input.View()
}

func (f Field) WithWidth(width int) Field {
	f.input.SetWidth(max(1, width-input.CursorWidth))
	f.input.SetCursor(f.input.Position())

	return f
}

func (f Field) ShortHelp() []key.Binding {
	return f.keys.ShortHelp()
}

func (f Field) FullHelp() [][]key.Binding {
	return f.keys.FullHelp()
}

func (f Field) committed() Result {
	name := f.input.Value()

	err := f.validate(name)
	if err != nil {
		return refusedFor(err)
	}

	return committedAs(name)
}
