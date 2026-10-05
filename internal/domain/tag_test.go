package domain_test

import (
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestTag_New(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        domain.TagID
		updatedAt time.Time
		wantErrs  []error
	}{
		{name: "accepts a Tag", id: tagID(), updatedAt: created},
		{name: "rejects the nil id", id: domain.TagID{}, updatedAt: created, wantErrs: []error{domain.ErrNilID}},
		{
			name:      "reports every broken rule",
			id:        domain.TagID{},
			updatedAt: created.Add(-time.Second),
			wantErrs:  []error{domain.ErrNilID, domain.ErrUpdatedBeforeCreate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewTag(tt.id, mustTagName(t, "go"), created, tt.updatedAt)

			requireErrors(t, err, tt.wantErrs)
		})
	}
}

func TestTag_Accessors(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	name := mustTagName(t, "Go")

	tag, err := domain.NewTag(tagID(), name, created, updated)

	require.NoError(t, err)
	assert.Equal(t, tagID(), tag.ID())
	assert.Equal(t, name, tag.Name())
	assert.Equal(t, created, tag.CreatedAt())
	assert.Equal(t, updated, tag.UpdatedAt())
}

func TestCompareTags(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})
	upperGo := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Go"})
	lowerGo := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"})
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Shell"})

	tags := []domain.Tag{shell, lowerGo, docker, upperGo}
	slices.SortFunc(tags, domain.CompareTags)

	assert.Equal(t, []domain.Tag{docker, upperGo, lowerGo, shell}, tags)
}

func tagID() domain.TagID {
	return domain.TagID(uuid.MustParse("0192f0c1-7a3b-7c4d-8e5f-0000000007a9"))
}

func mustTagName(tb testing.TB, raw string) value.TagName {
	tb.Helper()

	name, err := value.NewTagName(raw)
	require.NoError(tb, err)

	return name
}
