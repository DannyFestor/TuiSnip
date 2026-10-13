package domain

import (
	"cmp"
)

type SearchHit struct {
	snippet  Snippet
	fragment Fragment
	score    int
}

func NewSearchHit(snippet Snippet, scores FieldScores) SearchHit {
	best := bestWeightedFragment(snippet.FirstFragment(), scores.Fragments)
	fragment, fragmentScore := best.Fragment, weightedFragmentScore(best)

	snippetFieldScore := bestWeightedSnippetFieldScore(scores)
	if snippetFieldScore > fragmentScore {
		fragment = snippet.FirstFragment()
	}

	return SearchHit{snippet: snippet, fragment: fragment, score: max(snippetFieldScore, fragmentScore)}
}

func (h SearchHit) Snippet() Snippet {
	return h.snippet
}

func (h SearchHit) Fragment() Fragment {
	return h.fragment
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

func bestWeightedFragment(first Fragment, fragments []FragmentScores) FragmentScores {
	best := FragmentScores{Fragment: first, Title: 0, Content: 0}

	for index := range fragments {
		if weightedFragmentScore(fragments[index]) > weightedFragmentScore(best) {
			best = fragments[index]
		}
	}

	return best
}
