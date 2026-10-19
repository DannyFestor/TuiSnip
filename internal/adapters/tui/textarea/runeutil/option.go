package runeutil

type Option func(Sanitizer) Sanitizer

func ReplaceNewlines(nlRepl string) Option {
	return func(s Sanitizer) Sanitizer {
		s.replaceNewLine = []rune(nlRepl)

		return s
	}
}
