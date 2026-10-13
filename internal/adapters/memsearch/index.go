package memsearch

import (
	"context"
	"fmt"
	"slices"

	"github.com/junegunn/fzf/src/util"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
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

func (i *Index) Search(ctx context.Context, query value.SearchQuery) ([]domain.SearchHit, error) {
	snippets, err := i.lister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("memsearch.Index.Search: %w", err)
	}

	hits := matchAll(snippets, newPattern(query.String()))
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
	tags, tagMatched := tagScores(snippet.Tags(), query, slab)
	fragments, fragmentMatched := fragmentScores(snippet.Fragments(), query, slab)

	scores := domain.FieldScores{Title: title, Description: description, Tags: tags, Fragments: fragments}

	return scores, titleMatched || descriptionMatched || tagMatched || fragmentMatched
}

func fragmentScores(fragments []domain.Fragment, query pattern, slab *util.Slab) ([]domain.FragmentScores, bool) {
	scores := make([]domain.FragmentScores, 0, len(fragments))
	anyMatched := false

	for _, fragment := range fragments {
		title, titleMatched := query.fuzzyScore(fragment.Title().String(), slab)
		content, contentMatched := query.literalScore(fragment.Content().String(), slab)
		scores = append(scores, domain.FragmentScores{Fragment: fragment, Title: title, Content: content})
		anyMatched = anyMatched || titleMatched || contentMatched
	}

	return scores, anyMatched
}

func tagScores(tags []domain.Tag, query pattern, slab *util.Slab) ([]int, bool) {
	scores := make([]int, 0, len(tags))
	anyMatched := false

	for _, tag := range tags {
		score, matched := query.fuzzyScore(tag.Name().String(), slab)
		scores = append(scores, score)
		anyMatched = anyMatched || matched
	}

	return scores, anyMatched
}
