package runeutil

type Option func(Sanitizer) Sanitizer

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
