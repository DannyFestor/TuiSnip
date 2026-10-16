package runeutil

import (
	"unicode"
	"unicode/utf8"
)

type Sanitizer struct {
	replaceNewLine []rune
	replaceTab     []rune
}

type Option func(Sanitizer) Sanitizer

func NewSanitizer(opts ...Option) *Sanitizer {
	sanitizer := Sanitizer{
		replaceNewLine: []rune("\n"),
		replaceTab:     []rune("    "),
	}
	for _, option := range opts {
		sanitizer = option(sanitizer)
	}

	return &sanitizer
}

func ReplaceTabs(tabRepl string) Option {
	return func(s Sanitizer) Sanitizer {
		s.replaceTab = []rune(tabRepl)

		return s
	}
}

func ReplaceNewlines(nlRepl string) Option {
	return func(s Sanitizer) Sanitizer {
		s.replaceNewLine = []rune(nlRepl)

		return s
	}
}

// The rune array is modified in-place if possible. In that case, the
// returned slice is the original slice shortened after the control
// characters have been removed/translated.
func (s *Sanitizer) Sanitize(runes []rune) []rune {
	out := sanitizedRunes{runes: runes, dstrunes: runes[:0:len(runes)], copied: false}
	for src, char := range runes {
		s.write(&out, src, char)
	}

	return out.dstrunes
}

func (s *Sanitizer) write(out *sanitizedRunes, src int, char rune) {
	switch {
	case char == '\r' || char == '\n':
		out.replace(src, s.replaceNewLine)
	case char == '\t':
		out.replace(src, s.replaceTab)
	case char != utf8.RuneError && !unicode.IsControl(char):
		out.dstrunes = append(out.dstrunes, char)
	}
}

type sanitizedRunes struct {
	runes    []rune
	dstrunes []rune
	// copied indicates whether dstrunes is an alias of runes
	// or a copy. We need a copy when dst moves past src.
	// We use this as an optimization to avoid allocating
	// a new rune slice in the common case where the output
	// is smaller or equal to the input.
	copied bool
}

func (out *sanitizedRunes) replace(src int, replacement []rune) {
	if len(out.dstrunes)+len(replacement) > src && !out.copied {
		out.copyBeforeOverrun(len(replacement))
	}

	out.dstrunes = append(out.dstrunes, replacement...)
}

func (out *sanitizedRunes) copyBeforeOverrun(replacementLength int) {
	dst := len(out.dstrunes)
	out.dstrunes = make([]rune, dst, len(out.runes)+replacementLength)
	copy(out.dstrunes, out.runes[:dst])
	out.copied = true
}
