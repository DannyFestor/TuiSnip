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
	defaultSnippetID    = "0194c3a0-0000-7000-8000-000000005001"
	defaultSnippetTitle = "Snippet"
)

type SnippetSpec struct {
	ID          domain.SnippetID
	Title       string
	Description string
	FolderID    domain.FolderID
	Fragment    FragmentSpec
	Tags        []domain.Tag
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func Snippet(tb testing.TB, spec SnippetSpec) domain.Snippet {
	tb.Helper()

	title, err := value.NewTitle(cmp.Or(spec.Title, defaultSnippetTitle))
	require.NoError(tb, err)

	description, err := value.NewDescription(spec.Description)
	require.NoError(tb, err)

	createdAt := cmp.Or(spec.CreatedAt, defaultTime())
	snippet, err := domain.NewSnippet(
		cmp.Or(spec.ID, domain.SnippetID(uuid.MustParse(defaultSnippetID))),
		title,
		description,
		spec.FolderID,
		[]domain.Fragment{Fragment(tb, spec.Fragment)},
		spec.Tags,
		createdAt,
		cmp.Or(spec.UpdatedAt, createdAt),
	)
	require.NoError(tb, err)

	return snippet
}
