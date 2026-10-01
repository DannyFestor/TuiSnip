package value

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const maxTitleRunes = 200

var (
	ErrBlankTitle   = errors.New("value: title is blank")
	ErrTitleTooLong = errors.New("value: title is longer than 200 characters")
)

type Title struct{ value string }

func NewTitle(raw string) (Title, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Title{}, ErrBlankTitle
	}

	if utf8.RuneCountInString(trimmed) > maxTitleRunes {
		return Title{}, ErrTitleTooLong
	}

	return Title{value: trimmed}, nil
}

func (t Title) String() string {
	return t.value
}
