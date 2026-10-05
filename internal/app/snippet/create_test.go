package snippet_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const (
	maxDescriptionRunes = 2000
	maxContentBytes     = 256 * 1024
)

var errDatabaseLocked = errors.New("database is locked")

func TestNewCreate(t *testing.T) {
	t.Parallel()

	t.Run("names every missing dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewCreate(nil, nil, nil)

		require.ErrorIs(t, err, domain.ErrMissingDependency)
		require.ErrorContains(t, err, "inserter")
		require.ErrorContains(t, err, "ids")
		require.ErrorContains(t, err, "clock")
	})

	t.Run("accepts every dependency", func(t *testing.T) {
		t.Parallel()

		_, err := snippet.NewCreate(NewMockInserter(t), testkit.NewSequentialIDs(), fixedClock())

		assert.NoError(t, err)
	})
}

func TestCreate_Run(t *testing.T) {
	t.Parallel()

	t.Run("inserts a plain text Snippet at the Root", func(t *testing.T) {
		t.Parallel()

		var inserted domain.Snippet

		inserter := NewMockInserter(t)
		inserter.EXPECT().Insert(mock.Anything, mock.AnythingOfType("domain.Snippet")).
			Run(func(_ context.Context, s domain.Snippet) { inserted = s }).
			Return(nil)

		created, err := newCreate(t, inserter).Run(t.Context(), snippet.CreateInput{
			Title:       "  curl json ",
			Description: "POST with a JSON body",
			Language:    "plaintext",
			Content:     "curl -d @body.json\n",
		})

		require.NoError(t, err)
		assert.Equal(t, inserted, created)
		assertCreatedAtRoot(t, created)
	})

	t.Run("gives the Fragment the Language it was asked for", func(t *testing.T) {
		t.Parallel()

		inserter := NewMockInserter(t)
		inserter.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)

		created, err := newCreate(t, inserter).Run(t.Context(), snippet.CreateInput{
			Title: "curl json", Description: "", Language: "Bash", Content: "curl\n",
		})

		require.NoError(t, err)
		assert.Equal(t, "Bash", created.FirstFragment().Language().String())
	})

	t.Run("reports every invalid field at once", func(t *testing.T) {
		t.Parallel()

		_, err := newCreate(t, NewMockInserter(t)).Run(t.Context(), snippet.CreateInput{
			Title:       "   ",
			Description: strings.Repeat("d", maxDescriptionRunes+1),
			Language:    "golang",
			Content:     strings.Repeat("c", maxContentBytes+1),
		})

		require.ErrorIs(t, err, value.ErrBlankTitle)
		require.ErrorIs(t, err, value.ErrDescriptionTooLong)
		require.ErrorIs(t, err, value.ErrUnknownLanguage)
		require.ErrorIs(t, err, value.ErrContentTooLong)
		assert.Equal(
			t,
			[]domain.Field{domain.FieldTitle, domain.FieldDescription, domain.FieldLanguage, domain.FieldContent},
			fieldsOf(domain.FieldErrors(err)),
		)
	})

	t.Run("returns the insert error", func(t *testing.T) {
		t.Parallel()

		inserter := NewMockInserter(t)
		inserter.EXPECT().Insert(mock.Anything, mock.Anything).Return(errDatabaseLocked)

		_, err := newCreate(t, inserter).Run(
			t.Context(), snippet.CreateInput{Title: "curl json", Description: "", Language: "plaintext", Content: ""},
		)

		require.ErrorIs(t, err, errDatabaseLocked)
		assert.ErrorContains(t, err, "snippet.Create: ")
	})
}

func newCreate(t *testing.T, inserter snippet.Inserter) *snippet.Create {
	t.Helper()

	create, err := snippet.NewCreate(inserter, testkit.NewSequentialIDs(), fixedClock())
	require.NoError(t, err)

	return create
}

func assertCreatedAtRoot(t *testing.T, created domain.Snippet) {
	t.Helper()

	fragment := created.FirstFragment()
	assert.Equal(t, "curl json", created.Title().String())
	assert.Equal(t, "POST with a JSON body", created.Description().String())
	assert.Equal(t, "curl -d @body.json\n", fragment.Content().String())
	assert.Equal(t, value.PlainText(), fragment.Language())
	assert.True(t, created.AtRoot())
	assert.False(t, created.ID().IsNil())
	assert.False(t, fragment.ID().IsNil())
	assert.Equal(t, createdAt(), created.CreatedAt())
	assert.Equal(t, createdAt(), created.UpdatedAt())
	assert.Equal(t, createdAt(), fragment.CreatedAt())
}

func fieldsOf(fieldErrors []domain.FieldError) []domain.Field {
	fields := make([]domain.Field, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		fields = append(fields, fieldErr.Field)
	}

	return fields
}

func fixedClock() testkit.FixedClock {
	return testkit.NewFixedClock(createdAt())
}

func createdAt() time.Time {
	return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
}
