package value

import (
	"errors"
	"slices"
)

const plainTextName = "plaintext"

var ErrUnknownLanguage = errors.New("value: not a known language")

type Language struct{ name string }

func NewLanguage(raw string) (Language, error) {
	_, known := slices.BinarySearch(chromaLanguageNames(), raw)
	if !known {
		return Language{}, ErrUnknownLanguage
	}

	return Language{name: raw}, nil
}

func PlainText() Language {
	return Language{name: plainTextName}
}

func (l Language) String() string {
	return l.name
}
