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
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return trimmedName[K]{}, rules.errBlank
	}

	if utf8.RuneCountInString(trimmed) > rules.maxRunes {
		return trimmedName[K]{}, rules.errTooLong
	}

	return trimmedName[K]{value: trimmed}, nil
}

func (n trimmedName[K]) String() string {
	return n.value
}
