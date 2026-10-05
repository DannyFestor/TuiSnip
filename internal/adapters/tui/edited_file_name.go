package tui

import (
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	editedFileStem     = "snippet"
	plainEditedFile    = "snippet.txt"
	patternWildcard    = "*"
	patternMetacharset = "*?["
)

func editedFileName(language value.Language) string {
	lexer := lexers.Get(language.String())
	if lexer == nil || len(lexer.Config().Filenames) == 0 {
		return plainEditedFile
	}

	named := strings.Replace(lexer.Config().Filenames[0], patternWildcard, editedFileStem, 1)
	if strings.ContainsAny(named, patternMetacharset) {
		return plainEditedFile
	}

	return named
}
