package look

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const (
	DarkCodeStyle  = "github-dark"
	lightCodeStyle = "github"
	tabAsSpaces    = "    "
)

func Highlight(content, language, codeStyle string) string {
	plain := strings.ReplaceAll(strings.TrimSuffix(content, "\n"), "\t", tabAsSpaces)

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

func CodeStyleFor(background tea.BackgroundColorMsg) string {
	if background.IsDark() {
		return DarkCodeStyle
	}

	return lightCodeStyle
}
