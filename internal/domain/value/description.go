package value

import (
	"errors"
	"unicode/utf8"
)

const maxDescriptionRunes = 2000

var ErrDescriptionTooLong = errors.New("value: description is longer than 2000 characters")

type Description struct{ value string }

func NewDescription(raw string) (Description, error) {
	if utf8.RuneCountInString(raw) > maxDescriptionRunes {
		return Description{}, ErrDescriptionTooLong
	}

	return Description{value: raw}, nil
}

func (d Description) String() string {
	return d.value
}
