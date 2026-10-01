package config

import (
	"slices"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const (
	keyNameSpace     = "space"
	keyNameEscape    = "esc"
	writtenSpace     = " "
	writtenEscape    = "escape"
	functionKeyCount = 12
)

func namedKeys() []string {
	names := []string{
		"enter", keyNameEscape, keyNameSpace, "tab", "backspace", "delete", "insert",
		"up", "down", "left", "right", "home", "end", "pgup", "pgdown",
	}
	for number := 1; number <= functionKeyCount; number++ {
		names = append(names, "f"+strconv.Itoa(number))
	}

	return names
}

func isKeyName(name string) bool {
	return isPrintableCharacter(name) || slices.Contains(namedKeys(), name)
}

func isPrintableCharacter(name string) bool {
	character, size := utf8.DecodeRuneInString(name)

	return name != "" && size == len(name) && character != utf8.RuneError &&
		character != ' ' && unicode.IsPrint(character)
}

func reportedName(name string) string {
	switch name {
	case writtenSpace:
		return keyNameSpace
	case writtenEscape:
		return keyNameEscape
	}

	return name
}
