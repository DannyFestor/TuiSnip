package memsearch

import (
	"context"
	"fmt"
	"slices"

	"github.com/junegunn/fzf/src/util"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	slab16Size = 100 * 1024
	slab32Size = 2048
)

type Index struct {
	lister SnippetLister
}

func NewIndex(lister SnippetLister) *Index {
	initMatcher()

	return &Index{lister: lister}
}

func (i *Index) Search(ctx context.Context, query string) ([]domain.SearchHit, error) {
	snippets, err := i.lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("memsearch.Index.Search: %w", err)
	}

	hits := matchAll(snippets, newPattern(query))
	slices.SortFunc(hits, domain.CompareSearchHits)

	return hits, nil
}

func matchAll(snippets []domain.Snippet, query pattern) []domain.SearchHit {
	slab := util.MakeSlab(slab16Size, slab32Size)
	hits := make([]domain.SearchHit, 0, len(snippets))

	for i := range snippets {
		scores, matched := scoreFields(snippets[i], query, slab)
		if matched {
			hits = append(hits, domain.NewSearchHit(snippets[i], scores))
		}
	}

	return hits
}

func scoreFields(snippet domain.Snippet, query pattern, slab *util.Slab) (domain.FieldScores, bool) {
	title, titleMatched := query.fuzzyScore(snippet.Title().String(), slab)
	description, descriptionMatched := query.fuzzyScore(snippet.Description().String(), slab)
	content, contentMatched := query.literalScore(snippet.FirstFragment().Content().String(), slab)

	scores := domain.FieldScores{Title: title, Description: description, Tags: nil, Content: content}

	return scores, titleMatched || descriptionMatched || contentMatched
}
