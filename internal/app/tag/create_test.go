package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestNewCreate(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewCreate(nil, nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "tag.NewCreate: inserter")
		require.ErrorContains(t, err, "ids")
		assert.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := tag.NewCreate(NewMockInserter(t), testkit.NewSequentialIDs(), fixedClock())

		assert.NoError(t, err)
	})
}

func TestCreate_Run(t *testing.T) {
	t.Parallel()

	t.Run("inserts a Tag with the trimmed name, created now", func(t *testing.T) {
		t.Parallel()

		inserter := NewMockInserter(t)

		var inserted domain.Tag

		inserter.EXPECT().Insert(mock.Anything, mock.AnythingOfType("domain.Tag")).
			Run(func(_ context.Context, created domain.Tag) { inserted = created }).
			Return(nil)

		created, err := newCreate(t, inserter).Run(t.Context(), tag.CreateInput{Name: "  Go "})

		require.NoError(t, err)
		assert.Equal(t, inserted, created)
		assert.Equal(t, "Go", created.Name().String())
		assert.False(t, created.ID().IsNil())
		assert.Equal(t, now(), created.CreatedAt())
		assert.Equal(t, now(), created.UpdatedAt())
	})

	t.Run("refuses a name with a comma on the tag_name field", func(t *testing.T) {
		t.Parallel()

		_, err := newCreate(t, NewMockInserter(t)).Run(t.Context(), tag.CreateInput{Name: "go,rust"})

		require.ErrorIs(t, err, value.ErrTagNameHasComma)
		fieldErr, ok := errors.AsType[domain.FieldError](err)
		require.True(t, ok)
		assert.Equal(t, domain.FieldTagName, fieldErr.Field)
	})

	t.Run("passes up a name another Tag has", func(t *testing.T) {
		t.Parallel()

		inserter := NewMockInserter(t)
		inserter.EXPECT().Insert(mock.Anything, mock.Anything).Return(domain.ErrTagNameTaken)

		_, err := newCreate(t, inserter).Run(t.Context(), tag.CreateInput{Name: "go"})

		require.ErrorIs(t, err, domain.ErrTagNameTaken)
		assert.ErrorContains(t, err, "tag.Create")
	})
}

func newCreate(t *testing.T, inserter tag.Inserter) *tag.Create {
	t.Helper()

	create, err := tag.NewCreate(inserter, testkit.NewSequentialIDs(), fixedClock())
	require.NoError(t, err)

	return create
}
