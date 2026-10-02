package domain

import (
	"cmp"
)

type SearchHit struct {
	snippet Snippet
	score   int
}

func NewSearchHit(snippet Snippet, scores FieldScores) SearchHit {
	return SearchHit{snippet: snippet, score: bestWeightedScore(scores)}
}

func (h SearchHit) Snippet() Snippet {
	return h.snippet
}

func (h SearchHit) Score() int {
	return h.score
}

func CompareSearchHits(a, b SearchHit) int {
	return cmp.Or(
		cmp.Compare(b.score, a.score),
		b.snippet.UpdatedAt().Compare(a.snippet.UpdatedAt()),
		a.snippet.Title().CompareIgnoringCase(b.snippet.Title()),
		a.snippet.ID().Compare(b.snippet.ID()),
	)
}
