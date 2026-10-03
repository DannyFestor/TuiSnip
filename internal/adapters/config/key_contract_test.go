package config_test

import (
	"bytes"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/config"
)

const sampleModifiers = tea.ModCtrl | tea.ModAlt

type keystroke struct {
	modifiers tea.KeyMod
	name      string
}

type namedModifier struct {
	name     string
	modifier tea.KeyMod
}

func TestKey_String(t *testing.T) {
	t.Parallel()

	strokes := contractKeystrokes(t)

	for _, written := range slices.Sorted(maps.Keys(strokes)) {
		t.Run(written, func(t *testing.T) {
			t.Parallel()

			key, err := config.ParseKey(written)

			require.NoError(t, err)
			assert.Equal(t, strokes[written].pressed(t).String(), key.String())
		})
	}
}

func contractKeystrokes(t *testing.T) map[string]keystroke {
	t.Helper()

	strokes := make(map[string]keystroke)

	for _, written := range contractKeys(t) {
		plain := keystrokeOf(t, written)
		for _, stroke := range []keystroke{plain, plain.with(sampleModifiers)} {
			strokes[stroke.written()] = stroke
		}
	}

	return strokes
}

func contractKeys(t *testing.T) []string {
	t.Helper()

	cfg, err := config.Load(t.Context(), testOptions(filepath.Join(t.TempDir(), configFileName), &bytes.Buffer{}))
	require.NoError(t, err)

	keys := make(map[string]bool)
	for _, name := range config.AcceptedKeyNames() {
		keys[name] = true
	}

	for _, bindings := range cfg.Bindings {
		for _, bound := range bindings {
			for _, key := range bound {
				keys[key.String()] = true
			}
		}
	}

	return slices.Sorted(maps.Keys(keys))
}

func keystrokeOf(t *testing.T, written string) keystroke {
	t.Helper()

	separator := strings.LastIndex(written[:len(written)-1], "+")
	if separator < 0 {
		return keystroke{modifiers: 0, name: written}
	}

	var modifiers tea.KeyMod

	for name := range strings.SplitSeq(written[:separator], "+") {
		modifier, ok := modifierNamed(name)
		require.True(t, ok, "modifier %q", name)

		modifiers |= modifier
	}

	return keystroke{modifiers: modifiers, name: written[separator+1:]}
}

// Bubble Tea reports a capital typed with ctrl or alt as shift and the lower-case
// letter, so a config key for it is written that way.
func (k keystroke) with(modifiers tea.KeyMod) keystroke {
	k.modifiers |= modifiers

	character, size := utf8.DecodeRuneInString(k.name)
	if size == len(k.name) && unicode.IsUpper(character) {
		k.modifiers |= tea.ModShift
		k.name = string(unicode.ToLower(character))
	}

	return k
}

func (k keystroke) written() string {
	var written strings.Builder

	for _, modifier := range slices.Backward(modifierNames()) {
		if k.modifiers.Contains(modifier.modifier) {
			written.WriteString(modifier.name + "+")
		}
	}

	return written.String() + k.name
}

func (k keystroke) pressed(t *testing.T) tea.KeyPressMsg {
	t.Helper()

	if utf8.RuneCountInString(k.name) > 1 {
		code, ok := namedKeyCodes()[k.name]
		require.True(t, ok, "Bubble Tea code for %q", k.name)

		return tea.KeyPressMsg{Code: code, Mod: k.modifiers}
	}

	character, _ := utf8.DecodeRuneInString(k.name)
	press := tea.KeyPressMsg{Code: unicode.ToLower(character), Mod: k.modifiers}

	if unicode.IsUpper(character) {
		press.ShiftedCode = character
		press.Mod |= tea.ModShift
	}

	if press.Mod&^tea.ModShift == 0 {
		press.Text = k.name
	}

	return press
}

func modifierNames() []namedModifier {
	return []namedModifier{
		{name: "ctrl", modifier: tea.ModCtrl},
		{name: "alt", modifier: tea.ModAlt},
		{name: "shift", modifier: tea.ModShift},
		{name: "meta", modifier: tea.ModMeta},
		{name: "hyper", modifier: tea.ModHyper},
		{name: "super", modifier: tea.ModSuper},
	}
}

func modifierNamed(name string) (tea.KeyMod, bool) {
	for _, candidate := range modifierNames() {
		if candidate.name == name {
			return candidate.modifier, true
		}
	}

	return 0, false
}

func namedKeyCodes() map[string]rune {
	return map[string]rune{
		"enter": tea.KeyEnter, "esc": tea.KeyEscape, "space": tea.KeySpace, "tab": tea.KeyTab,
		"backspace": tea.KeyBackspace, "delete": tea.KeyDelete, "insert": tea.KeyInsert,
		"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
		"home": tea.KeyHome, "end": tea.KeyEnd, "pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown,
		"f1": tea.KeyF1, "f2": tea.KeyF2, "f3": tea.KeyF3, "f4": tea.KeyF4,
		"f5": tea.KeyF5, "f6": tea.KeyF6, "f7": tea.KeyF7, "f8": tea.KeyF8,
		"f9": tea.KeyF9, "f10": tea.KeyF10, "f11": tea.KeyF11, "f12": tea.KeyF12,
	}
}
