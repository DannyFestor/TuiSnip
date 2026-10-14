package domain

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

type FragmentLanguageFallback struct {
	Snippet  value.Title
	Position int
	Fragment value.FragmentTitle
	Lexer    string
}
