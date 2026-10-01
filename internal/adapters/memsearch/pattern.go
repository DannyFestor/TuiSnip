package memsearch

import (
	"strings"
	"sync"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

const (
	matcherScheme = "default"
	noMatchStart  = -1
	forward       = true
	withPositions = false
	foldsFuzzy    = true
	foldsLiteral  = false
)

//nolint:gochecknoglobals // fzf's tables are process-wide
var initMatcher = sync.OnceFunc(func() { algo.Init(matcherScheme) })

type pattern struct {
	fuzzy         []rune
	literal       []rune
	caseSensitive bool
}

func newPattern(query string) pattern {
	lowered := strings.ToLower(query)
	caseSensitive := lowered != query

	text := query
	if !caseSensitive {
		text = lowered
	}

	return pattern{
		fuzzy:         algo.NormalizeRunes([]rune(text)),
		literal:       []rune(text),
		caseSensitive: caseSensitive,
	}
}

func (p pattern) fuzzyScore(field string, slab *util.Slab) (int, bool) {
	input := util.ToChars([]byte(field))
	result, _ := algo.FuzzyMatchV2(p.caseSensitive, foldsFuzzy, forward, &input, p.fuzzy, withPositions, slab)

	return result.Score, result.Start != noMatchStart
}

func (p pattern) literalScore(field string, slab *util.Slab) (int, bool) {
	input := util.ToChars([]byte(field))
	result, _ := algo.ExactMatchNaive(p.caseSensitive, foldsLiteral, forward, &input, p.literal, withPositions, slab)

	return result.Score, result.Start != noMatchStart
}
