package value

import "errors"

const maxTitleRunes = 200

var (
	ErrBlankTitle   = errors.New("value: title is blank")
	ErrTitleTooLong = errors.New("value: title is longer than 200 characters")
)

type Title = trimmedName[titleKind]

func NewTitle(raw string) (Title, error) {
	return parseTrimmedName[titleKind](
		raw, nameRules{maxRunes: maxTitleRunes, errBlank: ErrBlankTitle, errTooLong: ErrTitleTooLong},
	)
}
