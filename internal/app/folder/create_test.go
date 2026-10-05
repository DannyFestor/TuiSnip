package folder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

var errDatabaseLocked = errors.New("database is locked")

func TestNewCreate(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewCreate(nil, nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "folder.NewCreate: repo")
		require.ErrorContains(t, err, "ids")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := folder.NewCreate(NewMockCreateRepository(t), testkit.NewSequentialIDs(), fixedClock())

		assert.NoError(t, err)
	})
}

func TestCreate_Run(t *testing.T) {
	t.Parallel()

	t.Run("inserts a plain text Folder at the Root", func(t *testing.T) {
		t.Parallel()

		repo := NewMockCreateRepository(t)
		inserted := expectInsert(repo)

		created, err := newCreate(
			t,
			repo,
		).Run(t.Context(), folder.CreateInput{Name: "  docker ", ParentID: domain.FolderID{}})

		require.NoError(t, err)
		assert.Equal(t, *inserted, created)
		assert.Equal(t, "docker", created.Name().String())
		assert.True(t, created.AtRoot())
		assert.Equal(t, value.PlainText(), created.DefaultLanguage())
		assert.False(t, created.ID().IsNil())
		assert.Equal(t, now(), created.CreatedAt())
	})

	t.Run("inserts a Folder inside its parent with the parent's Default Language", func(t *testing.T) {
		t.Parallel()

		parent := testkit.Folder(t, testkit.FolderSpec{Name: "go", DefaultLanguage: "Go"})
		repo := NewMockCreateRepository(t)
		repo.EXPECT().Find(mock.Anything, parent.ID()).Return(parent, nil)
		inserted := expectInsert(repo)

		created, err := newCreate(t, repo).Run(t.Context(), folder.CreateInput{Name: "testing", ParentID: parent.ID()})

		require.NoError(t, err)
		assert.Equal(t, *inserted, created)
		assert.Equal(t, parent.ID(), created.ParentID())
		assert.Equal(t, "Go", created.DefaultLanguage().String())
	})

	t.Run("refuses a blank name on the folder_name field", func(t *testing.T) {
		t.Parallel()

		_, err := newCreate(t, NewMockCreateRepository(t)).Run(t.Context(), folder.CreateInput{
			Name: "   ", ParentID: domain.FolderID{},
		})

		require.ErrorIs(t, err, value.ErrBlankFolderName)
		fieldErr, ok := errors.AsType[domain.FieldError](err)
		require.True(t, ok)
		assert.Equal(t, domain.FieldFolderName, fieldErr.Field)
	})

	t.Run("passes up a parent that is gone", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		repo := NewMockCreateRepository(t)
		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(domain.Folder{}, domain.ErrNotFound)

		_, err := newCreate(t, repo).Run(t.Context(), folder.CreateInput{Name: "testing", ParentID: ids.NewFolderID()})

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.ErrorContains(t, err, "folder.Create")
	})

	t.Run("passes up a failed insert", func(t *testing.T) {
		t.Parallel()

		repo := NewMockCreateRepository(t)
		repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newCreate(t, repo).Run(t.Context(), folder.CreateInput{Name: "docker", ParentID: domain.FolderID{}})

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "folder.Create")
	})
}

func newCreate(t *testing.T, repo folder.CreateRepository) *folder.Create {
	t.Helper()

	create, err := folder.NewCreate(repo, testkit.NewSequentialIDs(), fixedClock())
	require.NoError(t, err)

	return create
}

func expectInsert(repo *MockCreateRepository) *domain.Folder {
	var inserted domain.Folder

	repo.EXPECT().Insert(mock.Anything, mock.AnythingOfType("domain.Folder")).
		Run(func(_ context.Context, f domain.Folder) { inserted = f }).
		Return(nil)

	return &inserted
}
