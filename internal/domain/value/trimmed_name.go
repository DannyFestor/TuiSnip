package value

import (
	"strings"
	"unicode/utf8"
)

type trimmedName[K any] struct{ value string }

type (
	titleKind      struct{}
	folderNameKind struct{}
)

type nameRules struct {
	maxRunes   int
	errBlank   error
	errTooLong error
}

func parseTrimmedName[K any](raw string, rules nameRules) (trimmedName[K], error) {
	trimmed, err := trimmedWithin(raw, rules.maxRunes, rules.errTooLong)
	if err != nil {
		return trimmedName[K]{}, err
	}

	if trimmed == "" {
		return trimmedName[K]{}, rules.errBlank
	}

	return trimmedName[K]{value: trimmed}, nil
}

func trimmedWithin(raw string, maxRunes int, errTooLong error) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return "", errTooLong
	}

	return trimmed, nil
}

func (n trimmedName[K]) String() string {
	return n.value
}

// Only ASCII letters are folded, as SQLite's NOCASE does, so names order in Go the way Browse orders them.
func (n trimmedName[K]) CompareIgnoringCase(other trimmedName[K]) int {
	return strings.Compare(strings.Map(lowerASCIILetter, n.value), strings.Map(lowerASCIILetter, other.value))
}

func lowerASCIILetter(r rune) rune {
	if 'A' <= r && r <= 'Z' {
		return r + ('a' - 'A')
	}

	return r
}
