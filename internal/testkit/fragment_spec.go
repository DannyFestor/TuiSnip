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

const defaultFragmentID = "0194c3a0-0000-7000-8000-00000000f001"

type FragmentSpec struct {
	ID        domain.FragmentID
	Language  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Fragment(tb testing.TB, spec FragmentSpec) domain.Fragment {
	tb.Helper()

	language, err := value.NewLanguage(cmp.Or(spec.Language, value.PlainText().String()))
	require.NoError(tb, err)

	content, err := value.NewContent(spec.Content)
	require.NoError(tb, err)

	createdAt := cmp.Or(spec.CreatedAt, defaultTime())
	fragment, err := domain.NewFragment(
		cmp.Or(spec.ID, domain.FragmentID(uuid.MustParse(defaultFragmentID))),
		language,
		content,
		createdAt,
		cmp.Or(spec.UpdatedAt, createdAt),
	)
	require.NoError(tb, err)

	return fragment
}
