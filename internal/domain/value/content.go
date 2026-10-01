package value

import "errors"

// Highlighting and the textarea slow down on huge Fragments. The limit starts low because
// raising it later is cheap, while lowering it would strand Snippets already over it.
const maxContentBytes = 256 << 10

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
