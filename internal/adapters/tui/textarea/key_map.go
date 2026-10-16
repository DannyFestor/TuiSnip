package textarea

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	CharacterBackward       key.Binding
	CharacterForward        key.Binding
	DeleteAfterCursor       key.Binding
	DeleteBeforeCursor      key.Binding
	DeleteCharacterBackward key.Binding
	DeleteCharacterForward  key.Binding
	DeleteWordBackward      key.Binding
	DeleteWordForward       key.Binding
	InsertNewline           key.Binding
	LineEnd                 key.Binding
	LineNext                key.Binding
	LinePrevious            key.Binding
	LineStart               key.Binding
	PageUp                  key.Binding
	PageDown                key.Binding
	WordBackward            key.Binding
	WordForward             key.Binding
	InputBegin              key.Binding
	InputEnd                key.Binding

	UppercaseWordForward  key.Binding
	LowercaseWordForward  key.Binding
	CapitalizeWordForward key.Binding

	TransposeCharacterBackward key.Binding

	SelectCharacterForward  key.Binding
	SelectCharacterBackward key.Binding
	SelectWordForward       key.Binding
	SelectWordBackward      key.Binding
	SelectLineUp            key.Binding
	SelectLineDown          key.Binding
	SelectAll               key.Binding
}

func DefaultKeyMap() KeyMap {
	var keyMap KeyMap

	keyMap.bindMovement()
	keyMap.bindDeletion()
	keyMap.bindEditing()
	keyMap.bindSelection()

	return keyMap
}

func (k *KeyMap) bindMovement() {
	k.CharacterForward = key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("right", "character forward"))
	k.CharacterBackward = key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("left", "character backward"))
	k.WordForward = key.NewBinding(
		key.WithKeys("alt+right", "ctrl+right", "alt+f"),
		key.WithHelp("alt+right", "word forward"),
	)
	k.WordBackward = key.NewBinding(
		key.WithKeys("alt+left", "ctrl+left", "alt+b"),
		key.WithHelp("alt+left", "word backward"),
	)
	k.LineNext = key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "next line"))
	k.LinePrevious = key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "previous line"))
	k.LineStart = key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "line start"))
	k.LineEnd = key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "line end"))
	k.PageUp = key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up"))
	k.PageDown = key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down"))
	k.InputBegin = key.NewBinding(key.WithKeys("alt+<", "ctrl+home"), key.WithHelp("alt+<", "input begin"))
	k.InputEnd = key.NewBinding(key.WithKeys("alt+>", "ctrl+end"), key.WithHelp("alt+>", "input end"))
}

func (k *KeyMap) bindDeletion() {
	k.DeleteWordBackward = key.NewBinding(
		key.WithKeys("alt+backspace", "ctrl+w", "ctrl+backspace"),
		key.WithHelp("alt+backspace", "delete word backward"),
	)
	k.DeleteWordForward = key.NewBinding(
		key.WithKeys("alt+delete", "alt+d", "ctrl+delete"),
		key.WithHelp("alt+delete", "delete word forward"),
	)
	k.DeleteAfterCursor = key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete after cursor"))
	k.DeleteBeforeCursor = key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "delete before cursor"))
	k.DeleteCharacterBackward = key.NewBinding(
		key.WithKeys("backspace", "ctrl+h"),
		key.WithHelp("backspace", "delete character backward"),
	)
	k.DeleteCharacterForward = key.NewBinding(
		key.WithKeys("delete", "ctrl+d"),
		key.WithHelp("delete", "delete character forward"),
	)
}

func (k *KeyMap) bindEditing() {
	k.InsertNewline = key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "insert newline"))
	k.CapitalizeWordForward = key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt+c", "capitalize word forward"))
	k.LowercaseWordForward = key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "lowercase word forward"))
	k.UppercaseWordForward = key.NewBinding(key.WithKeys("alt+u"), key.WithHelp("alt+u", "uppercase word forward"))
	k.TransposeCharacterBackward = key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "transpose character backward"),
	)
}

func (k *KeyMap) bindSelection() {
	k.SelectCharacterForward = key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+right", "select character forward"),
	)
	k.SelectCharacterBackward = key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+left", "select character backward"),
	)
	k.SelectWordForward = key.NewBinding(
		key.WithKeys("ctrl+shift+right", "alt+shift+right", "alt+shift+f"),
		key.WithHelp("alt+shift+right", "select word forward"),
	)
	k.SelectWordBackward = key.NewBinding(
		key.WithKeys("ctrl+shift+left", "alt+shift+left", "alt+shift+b"),
		key.WithHelp("alt+shift+left", "select word backward"),
	)
	k.SelectLineUp = key.NewBinding(key.WithKeys("shift+up"), key.WithHelp("shift+up", "select line up"))
	k.SelectLineDown = key.NewBinding(key.WithKeys("shift+down"), key.WithHelp("shift+down", "select line down"))
	k.SelectAll = key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "select all"))
}
