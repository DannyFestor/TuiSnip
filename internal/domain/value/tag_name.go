package value

import (
	"errors"
	"strings"
	"unicode"
)

const tagNameSeparator = ","

var (
	ErrBlankTagName    = errors.New("value: tag name is blank")
	ErrTagNameHasComma = errors.New("value: tag name contains a comma")
)

type TagName struct{ value string }

func NewTagName(raw string) (TagName, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return TagName{}, ErrBlankTagName
	}

	if strings.Contains(trimmed, tagNameSeparator) {
		return TagName{}, ErrTagNameHasComma
	}

	return TagName{value: trimmed}, nil
}

func (n TagName) String() string {
	return n.value
}

func (n TagName) Key() string {
	return foldedKey(n.value)
}

func TagNameKey(raw string) string {
	return foldedKey(strings.TrimSpace(raw))
}

func foldedKey(text string) string {
	return strings.Map(smallestOfFoldOrbit, text)
}

func smallestOfFoldOrbit(r rune) rune {
	smallest := r
	for folded := unicode.SimpleFold(r); folded != r; folded = unicode.SimpleFold(folded) {
		smallest = min(smallest, folded)
	}

	return smallest
}
