package testkit

import (
	"cmp"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	defaultTagID   = "0194c3a0-0000-7000-8000-0000000007a9"
	defaultTagName = "Tag"
)

type TagSpec struct {
	ID        domain.TagID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Tag(tb testing.TB, spec TagSpec) domain.Tag {
	tb.Helper()

	name, err := value.NewTagName(cmp.Or(spec.Name, defaultTagName))
	require.NoError(tb, err)

	createdAt := cmp.Or(spec.CreatedAt, defaultTime())
	tag, err := domain.NewTag(
		cmp.Or(spec.ID, domain.TagID(uuid.MustParse(defaultTagID))),
		name,
		createdAt,
		cmp.Or(spec.UpdatedAt, createdAt),
	)
	require.NoError(tb, err)

	return tag
}
