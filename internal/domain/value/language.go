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

func Languages() []Language {
	names := chromaLanguageNames()
	languages := make([]Language, 0, len(names))

	for _, name := range names {
		languages = append(languages, Language{name: name})
	}

	return languages
}

func PlainText() Language {
	return Language{name: plainTextName}
}

func (l Language) String() string {
	return l.name
}
