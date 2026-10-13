package value

import (
	"errors"
)

const maxFragmentTitleRunes = 200

var ErrFragmentTitleTooLong = errors.New("value: fragment title is longer than 200 characters")

type FragmentTitle struct{ value string }

func NewFragmentTitle(raw string) (FragmentTitle, error) {
	trimmed, err := trimmedWithin(raw, maxFragmentTitleRunes, ErrFragmentTitleTooLong)
	if err != nil {
		return FragmentTitle{}, err
	}

	return FragmentTitle{value: trimmed}, nil
}

func (t FragmentTitle) String() string {
	return t.value
}

func (t FragmentTitle) IsSet() bool {
	return t.value != ""
}
