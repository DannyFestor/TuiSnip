package snippetslab

import (
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type fragmentPlace struct {
	snippet  value.Title
	position int
}

func (p fragmentPlace) fellBack(title value.FragmentTitle, lexer string) domain.FragmentLanguageFallback {
	return domain.FragmentLanguageFallback{Snippet: p.snippet, Position: p.position, Fragment: title, Lexer: lexer}
}
