package value

import (
	"errors"
	"strings"
)

// Highlighting and the textarea slow down on huge Fragments. The limit starts low because
// raising it later is cheap, while lowering it would strand Snippets already over it.
const maxContentBytes = 256 << 10

const (
	windowsLineEnding = "\r\n"
	unixLineEnding    = "\n"
)

var ErrContentTooLong = errors.New("value: content is larger than 256 KiB")

type Content struct{ value string }

func NewContent(raw string) (Content, error) {
	if len(raw) > maxContentBytes {
		return Content{}, ErrContentTooLong
	}

	return Content{value: raw}, nil
}

func (c Content) String() string {
	return c.value
}

func (c Content) WithoutTrailingNewline() Content {
	if trimmed, cut := strings.CutSuffix(c.value, windowsLineEnding); cut {
		return Content{value: trimmed}
	}

	return Content{value: strings.TrimSuffix(c.value, unixLineEnding)}
}
