package picker_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 40
	noMatches = "Nothing matches."
)

func opened(t *testing.T, choices []picker.Choice) picker.List {
	t.Helper()

	return openedWith(t, testsettings.Default(t).Keys, choices)
}

func openedWith(t *testing.T, keys binding.Keys, choices []picker.Choice) picker.List {
	t.Helper()

	list, _ := picker.New(keys, picker.Labels{Prompt: "filter: ", NoMatches: noMatches}, choices)

	return resized(list, look.Size{Width: boxWidth, Height: 20})
}

func resized(list picker.List, box look.Size) picker.List {
	next, _, _ := list.Update(look.Resized{Box: box})

	return next
}

func choicesOf(texts ...string) []picker.Choice {
	choices := make([]picker.Choice, 0, len(texts))
	for _, text := range texts {
		choices = append(choices, picker.Choice{Text: text, Meta: ""})
	}

	return choices
}

func press(t *testing.T, list picker.List, key tea.KeyPressMsg) (picker.List, picker.Result) {
	t.Helper()

	next, result, _ := list.Update(key)

	return next, result
}

func pressed(t *testing.T, list picker.List, keys ...tea.KeyPressMsg) picker.List {
	t.Helper()

	for _, key := range keys {
		list, _ = press(t, list, key)
	}

	return list
}

func typed(t *testing.T, list picker.List, text string) picker.List {
	t.Helper()

	return pressed(t, list, keypress.Typed(text)...)
}

func picked(index int) picker.Result {
	return picker.Result{Ending: picker.Picked, Index: index}
}

func down() tea.KeyPressMsg {
	return keypress.Special(tea.KeyDown)
}

func up() tea.KeyPressMsg {
	return keypress.Special(tea.KeyUp)
}

func enter() tea.KeyPressMsg {
	return keypress.Special(tea.KeyEnter)
}
