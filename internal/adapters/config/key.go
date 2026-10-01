package config

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Key struct {
	modifiers modifiers
	name      string
}

func ParseKey(written string) (Key, error) {
	key, ok := splitKey(written)
	if !ok {
		return Key{}, fmt.Errorf("%q %w", written, errNotAKey)
	}

	reported, err := key.asBubbleTeaReportsIt()
	if err != nil {
		return Key{}, fmt.Errorf("%q %w", written, err)
	}

	if !isKeyName(reported.name) {
		return Key{}, fmt.Errorf("%q %w", written, errNotAKey)
	}

	if reported != key {
		return Key{}, fmt.Errorf("%q %w; write %q", written, errNeverMatches, reported.String())
	}

	return key, nil
}

func fixedQuitKey() Key {
	return Key{modifiers: modifierCtrl, name: "c"}
}

func splitKey(written string) (Key, bool) {
	if written == "" {
		return Key{}, false
	}

	separator := strings.LastIndex(written[:len(written)-1], modifierSeparator)
	if separator < 0 {
		return Key{modifiers: 0, name: written}, true
	}

	parsed, ok := parseModifiers(written[:separator])

	return Key{modifiers: parsed, name: written[separator+1:]}, ok
}

func (k Key) String() string {
	return k.modifiers.prefix() + k.name
}

func (k Key) typesText() bool {
	return k.modifiers == 0 && (k.name == keyNameSpace || isPrintableCharacter(k.name))
}

func (k Key) asBubbleTeaReportsIt() (Key, error) {
	k.name = reportedName(k.name)
	if !isPrintableCharacter(k.name) {
		return k, nil
	}

	if k.modifiers == modifierShift {
		return k.asTypedCharacter()
	}

	return k.withShiftSpelledOut(), nil
}

func (k Key) asTypedCharacter() (Key, error) {
	character, _ := utf8.DecodeRuneInString(k.name)

	upper := unicode.ToUpper(character)
	if !unicode.IsUpper(upper) {
		return Key{}, errShiftedSymbol
	}

	return Key{modifiers: 0, name: string(upper)}, nil
}

func (k Key) withShiftSpelledOut() Key {
	character, _ := utf8.DecodeRuneInString(k.name)
	if k.modifiers == 0 || !unicode.IsUpper(character) {
		return k
	}

	return Key{modifiers: k.modifiers | modifierShift, name: string(unicode.ToLower(character))}
}
