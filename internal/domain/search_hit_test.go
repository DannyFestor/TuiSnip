package domain_test

import (
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestSearchHit_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		scores domain.FieldScores
		want   int
	}{
		{name: "weights title by 4", scores: domain.FieldScores{Title: 10}, want: 40},
		{name: "weights a Tag by 3", scores: domain.FieldScores{Tags: []int{10}}, want: 30},
		{name: "weights Description by 2", scores: domain.FieldScores{Description: 10}, want: 20},
		{
			name:   "weights a Fragment title by 3",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{Title: 10}}},
			want:   30,
		},
		{
			name:   "weights content by 1",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{Content: 10}}},
			want:   10,
		},
		{
			name: "takes the best weighted field",
			scores: domain.FieldScores{
				Title:       5,
				Description: 11,
				Tags:        []int{2, 7},
				Fragments:   []domain.FragmentScores{{Title: 3, Content: 30}},
			},
			want: 30,
		},
		{name: "takes the best of several Tags", scores: domain.FieldScores{Tags: []int{4, 9, 1}}, want: 27},
		{
			name:   "takes the best of several Fragments",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{Content: 4}, {Title: 2}, {Content: 1}}},
			want:   6,
		},
		{name: "scores zero when no field matched", scores: domain.FieldScores{}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hit := domain.NewSearchHit(testkit.Snippet(t, testkit.SnippetSpec{}), tt.scores)

			assert.Equal(t, tt.want, hit.Score())
		})
	}
}

func TestSearchHit_Snippet(t *testing.T) {
	t.Parallel()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{Title: "curl json"})

	hit := domain.NewSearchHit(snippet, domain.FieldScores{Title: 1})

	assert.Equal(t, snippet, hit.Snippet())
}

func TestSearchHit_Fragment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		scores domain.FieldScores
		want   int
	}{
		{
			name:   "opens on the Fragment whose content matched",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{}, {}, {Content: 5}}},
			want:   2,
		},
		{
			name:   "opens on the Fragment whose title matched",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{}, {Title: 5}, {}}},
			want:   1,
		},
		{
			name:   "opens on the best weighted Fragment",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{Content: 8}, {Title: 3}, {}}},
			want:   1,
		},
		{
			name:   "opens on the earliest of tied Fragments",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{}, {Content: 6}, {Title: 2}}},
			want:   1,
		},
		{
			name: "opens on the first Fragment when a Snippet field matched best",
			scores: domain.FieldScores{
				Description: 5, Fragments: []domain.FragmentScores{{}, {Content: 9}, {}},
			},
			want: 0,
		},
		{
			name: "opens on a Fragment that ties with a Snippet field",
			scores: domain.FieldScores{
				Title: 1, Fragments: []domain.FragmentScores{{}, {}, {Content: 4}},
			},
			want: 2,
		},
		{
			name:   "opens on the first Fragment when nothing matched",
			scores: domain.FieldScores{Fragments: []domain.FragmentScores{{}, {}, {}}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippet := snippetWithFragments(t, 3)

			hit := domain.NewSearchHit(snippet, scoringFragmentsOf(snippet, tt.scores))

			assert.Equal(t, snippet.Fragments()[tt.want], hit.Fragment())
		})
	}
}

func TestCompareSearchHits(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	first := domain.SnippetID(uuid.MustParse("0194c3a0-0000-7000-8000-000000000001"))
	second := domain.SnippetID(uuid.MustParse("0194c3a0-0000-7000-8000-000000000002"))

	tests := []struct {
		name      string
		hits      []searchHitSpec
		wantOrder []domain.SnippetID
	}{
		{
			name:      "orders by score, highest first",
			hits:      []searchHitSpec{{id: first, title: "low", score: 1}, {id: second, title: "high", score: 9}},
			wantOrder: []domain.SnippetID{second, first},
		},
		{
			name: "breaks a score tie by the most recently updated",
			hits: []searchHitSpec{
				{id: first, title: "older", score: 5, updatedAt: older},
				{id: second, title: "newer", score: 5, updatedAt: newer},
			},
			wantOrder: []domain.SnippetID{second, first},
		},
		{
			name: "breaks a full tie by title",
			hits: []searchHitSpec{
				{id: first, title: "zsh", score: 5, updatedAt: older},
				{id: second, title: "awk", score: 5, updatedAt: older},
			},
			wantOrder: []domain.SnippetID{second, first},
		},
		{
			name: "orders titles case-insensitively",
			hits: []searchHitSpec{
				{id: first, title: "Bash", score: 5, updatedAt: older},
				{id: second, title: "awk", score: 5, updatedAt: older},
			},
			wantOrder: []domain.SnippetID{second, first},
		},
		{
			name: "breaks a tie on titles differing only in case by ID",
			hits: []searchHitSpec{
				{id: second, title: "Curl", score: 5, updatedAt: older},
				{id: first, title: "curl", score: 5, updatedAt: older},
			},
			wantOrder: []domain.SnippetID{first, second},
		},
		{
			name: "breaks a tie on identical titles by ID",
			hits: []searchHitSpec{
				{id: second, title: "curl", score: 5, updatedAt: older},
				{id: first, title: "curl", score: 5, updatedAt: older},
			},
			wantOrder: []domain.SnippetID{first, second},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hits := buildSearchHits(t, tt.hits)

			slices.SortFunc(hits, domain.CompareSearchHits)

			assert.Equal(t, tt.wantOrder, hitIDs(hits))
		})
	}
}

type searchHitSpec struct {
	id        domain.SnippetID
	title     string
	score     int
	updatedAt time.Time
}

func buildSearchHits(t *testing.T, specs []searchHitSpec) []domain.SearchHit {
	t.Helper()

	hits := make([]domain.SearchHit, 0, len(specs))
	for _, spec := range specs {
		snippet := testkit.Snippet(t, testkit.SnippetSpec{ID: spec.id, Title: spec.title, UpdatedAt: spec.updatedAt})
		hits = append(hits, domain.NewSearchHit(snippet, domain.FieldScores{Title: spec.score}))
	}

	return hits
}

func snippetWithFragments(t *testing.T, count int) domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	fragments := make([]testkit.FragmentSpec, 0, count)

	for range count {
		fragments = append(fragments, testkit.FragmentSpec{ID: ids.NewFragmentID()})
	}

	return testkit.Snippet(t, testkit.SnippetSpec{Fragments: fragments})
}

func scoringFragmentsOf(snippet domain.Snippet, scores domain.FieldScores) domain.FieldScores {
	scored := scores
	scored.Fragments = slices.Clone(scores.Fragments)

	for index := range scored.Fragments {
		scored.Fragments[index].Fragment = snippet.Fragments()[index]
	}

	return scored
}

func hitIDs(hits []domain.SearchHit) []domain.SnippetID {
	ids := make([]domain.SnippetID, 0, len(hits))
	for i := range hits {
		ids = append(ids, hits[i].Snippet().ID())
	}

	return ids
}
