package main

import (
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// highlightModes: "background" lexes the whole Fragment off the UI thread once typing pauses;
// "document" lexes it on the UI thread at every change; both serve cached ranges. "line" lexes
// each line on its own at every render; "off" draws plain text.
var highlightModes = []string{"background", "document", "line", "off"}

type highlighter struct {
	lexer  chroma.Lexer
	style  *chroma.Style
	mode   string
	cached [][]lipgloss.Range
	lexed  string
	looks  map[chroma.TokenType]lipgloss.Style

	looksLock sync.Mutex

	lineLexTime time.Duration
}

func newHighlighter(language string) *highlighter {
	h := &highlighter{style: styles.Get("github-dark"), mode: highlightModes[0], looks: map[chroma.TokenType]lipgloss.Style{}}
	h.setLanguage(language)
	return h
}

func (h *highlighter) setLanguage(language string) {
	h.lexer = lexers.Get(language)
	if h.lexer == nil {
		h.lexer = lexers.Fallback
	}
	h.lexer = chroma.Coalesce(h.lexer)
}

func (h *highlighter) language() string {
	return h.lexer.Config().Name
}

func (h *highlighter) Highlight(lineIdx int, line []rune) []lipgloss.Range {
	switch h.mode {
	case "background", "document":
		if lineIdx < len(h.cached) {
			return h.cached[lineIdx]
		}
	case "line":
		start := time.Now()
		ranges := h.lex(string(line))
		h.lineLexTime += time.Since(start)
		if len(ranges) > 0 {
			return ranges[0]
		}
	}
	return nil
}

func (h *highlighter) relex(document string) time.Duration {
	start := time.Now()
	h.cached = h.lex(document)
	h.lexed = document
	return time.Since(start)
}

func (h *highlighter) lex(text string) [][]lipgloss.Range {
	iterator, err := h.lexer.Tokenise(nil, text)
	if err != nil {
		return nil
	}
	ranges := [][]lipgloss.Range{nil}
	col := 0
	for token := iterator(); token != chroma.EOF; token = iterator() {
		for i, part := range strings.Split(token.Value, "\n") {
			if i > 0 {
				ranges = append(ranges, nil)
				col = 0
			}
			width := utf8.RuneCountInString(part)
			if width == 0 {
				continue
			}
			row := len(ranges) - 1
			ranges[row] = append(ranges[row], lipgloss.NewRange(col, col+width, h.look(token.Type)))
			col += width
		}
	}
	return ranges
}

func (h *highlighter) look(tokenType chroma.TokenType) lipgloss.Style {
	h.looksLock.Lock()
	defer h.looksLock.Unlock()
	if cached, ok := h.looks[tokenType]; ok {
		return cached
	}
	entry := h.style.Get(tokenType)
	look := lipgloss.NewStyle()
	if entry.Colour.IsSet() {
		look = look.Foreground(lipgloss.Color(entry.Colour.String()))
	}
	look = look.Bold(entry.Bold == chroma.Yes).Italic(entry.Italic == chroma.Yes)
	h.looks[tokenType] = look
	return look
}
