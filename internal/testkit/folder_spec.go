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
	defaultFolderID   = "0194c3a0-0000-7000-8000-00000000f01d"
	defaultFolderName = "Folder"
)

type FolderSpec struct {
	ID              domain.FolderID
	Name            string
	ParentID        domain.FolderID
	DefaultLanguage string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func Folder(tb testing.TB, spec FolderSpec) domain.Folder {
	tb.Helper()

	name, err := value.NewFolderName(cmp.Or(spec.Name, defaultFolderName))
	require.NoError(tb, err)

	language, err := value.NewLanguage(cmp.Or(spec.DefaultLanguage, value.PlainText().String()))
	require.NoError(tb, err)

	createdAt := cmp.Or(spec.CreatedAt, defaultTime())
	folder, err := domain.NewFolder(
		cmp.Or(spec.ID, domain.FolderID(uuid.MustParse(defaultFolderID))),
		name,
		spec.ParentID,
		language,
		createdAt,
		cmp.Or(spec.UpdatedAt, createdAt),
	)
	require.NoError(tb, err)

	return folder
}
