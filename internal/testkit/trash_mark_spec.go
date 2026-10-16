package testkit

import (
	"cmp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type TrashMarkSpec struct {
	Item       domain.TrashItem
	DeletedAt  time.Time
	FormerPath []string
}

func TrashMark(tb testing.TB, spec TrashMarkSpec) domain.TrashMark {
	tb.Helper()

	formerPath, err := value.NewFolderPath(spec.FormerPath)
	require.NoError(tb, err)

	mark, err := domain.NewTrashMark(spec.Item, cmp.Or(spec.DeletedAt, spec.Item.CreatedAt()), formerPath)
	require.NoError(tb, err)

	return mark
}
