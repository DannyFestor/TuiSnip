package look

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
)

func Highlight(content, language, codeStyle string) string {
	plain := tabstop.Expand(strings.TrimSuffix(content, "\n"))

	iterator, err := tokenise(plain, language)
	if err != nil {
		return plain
	}

	var highlighted strings.Builder

	err = formatters.TTY16m.Format(&highlighted, styles.Get(codeStyle), iterator)
	if err != nil {
		return plain
	}

	return highlighted.String()
}

func tokenise(plain, language string) (chroma.Iterator, error) {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, plain)
	if err != nil {
		return nil, fmt.Errorf("tokenise %s: %w", language, err)
	}

	return iterator, nil
}
