package domain_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestFragment_New(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        domain.FragmentID
		createdAt time.Time
		updatedAt time.Time
		wantErrs  []error
	}{
		{name: "accepts equal timestamps", id: fragmentID(), createdAt: created, updatedAt: created},
		{name: "accepts a later update", id: fragmentID(), createdAt: created, updatedAt: created.Add(time.Second)},
		{
			name:      "rejects the nil id",
			id:        domain.FragmentID{},
			createdAt: created,
			updatedAt: created,
			wantErrs:  []error{domain.ErrNilID},
		},
		{
			name:      "rejects an update before creation",
			id:        fragmentID(),
			createdAt: created,
			updatedAt: created.Add(-time.Nanosecond),
			wantErrs:  []error{domain.ErrUpdatedBeforeCreate},
		},
		{
			name:      "reports every broken rule",
			id:        domain.FragmentID{},
			createdAt: time.Time{},
			updatedAt: created,
			wantErrs:  []error{domain.ErrNilID, domain.ErrTimestampOutOfRange},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewFragment(tt.id, value.PlainText(), mustContent(t, "ls -la"), tt.createdAt, tt.updatedAt)

			requireErrors(t, err, tt.wantErrs)
		})
	}
}

func TestFragment_Accessors(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	content := mustContent(t, "go test ./...")

	fragment, err := domain.NewFragment(fragmentID(), value.PlainText(), content, created, updated)

	require.NoError(t, err)
	assert.Equal(t, fragmentID(), fragment.ID())
	assert.Equal(t, value.PlainText(), fragment.Language())
	assert.Equal(t, content, fragment.Content())
	assert.Equal(t, created, fragment.CreatedAt())
	assert.Equal(t, updated, fragment.UpdatedAt())
}

func TestFragment_Edit(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	later := created.Add(time.Hour)
	fragment := testkit.Fragment(
		t,
		testkit.FragmentSpec{ID: fragmentID(), Language: "Bash", Content: "ls", CreatedAt: created},
	)

	t.Run("takes new content and the time", func(t *testing.T) {
		t.Parallel()

		edited, err := fragment.Edit(mustLanguage(t, "Bash"), mustContent(t, "ls -la"), later)

		require.NoError(t, err)
		assert.Equal(t, "ls -la", edited.Content().String())
		assert.Equal(t, later, edited.UpdatedAt())
		assert.Equal(t, "ls", fragment.Content().String(), "the original is untouched")
	})

	t.Run("takes a new Language and the time", func(t *testing.T) {
		t.Parallel()

		edited, err := fragment.Edit(mustLanguage(t, "Python"), mustContent(t, "ls"), later)

		require.NoError(t, err)
		assert.Equal(t, "Python", edited.Language().String())
		assert.Equal(t, later, edited.UpdatedAt())
		assert.Equal(t, "Bash", fragment.Language().String(), "the original is untouched")
	})

	t.Run("keeps its time when the Language and content are the same", func(t *testing.T) {
		t.Parallel()

		edited, err := fragment.Edit(mustLanguage(t, "Bash"), mustContent(t, "ls"), later)

		require.NoError(t, err)
		assert.Equal(t, created, edited.UpdatedAt())
	})

	t.Run("rejects a time before creation", func(t *testing.T) {
		t.Parallel()

		_, err := fragment.Edit(mustLanguage(t, "Bash"), mustContent(t, "ls -la"), created.Add(-time.Hour))

		require.ErrorIs(t, err, domain.ErrUpdatedBeforeCreate)
	})
}

func TestFragment_Duplicate(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	later := created.Add(time.Hour)
	duplicateID := testkit.NewSequentialIDs().NewFragmentID()
	fragment := testkit.Fragment(
		t,
		testkit.FragmentSpec{ID: fragmentID(), Language: "Bash", Content: "ls", CreatedAt: created},
	)

	t.Run("keeps the Language and content under the new id and time", func(t *testing.T) {
		t.Parallel()

		duplicate, err := fragment.Duplicate(duplicateID, later)

		require.NoError(t, err)
		assert.Equal(t, duplicateID, duplicate.ID())
		assert.Equal(t, fragment.Language(), duplicate.Language())
		assert.Equal(t, fragment.Content(), duplicate.Content())
		assert.Equal(t, later, duplicate.CreatedAt())
		assert.Equal(t, later, duplicate.UpdatedAt())
	})

	t.Run("rejects the nil id", func(t *testing.T) {
		t.Parallel()

		_, err := fragment.Duplicate(domain.FragmentID{}, later)

		require.ErrorIs(t, err, domain.ErrNilID)
	})
}

func fragmentID() domain.FragmentID {
	return domain.FragmentID(uuid.MustParse(storedID))
}

func mustContent(t *testing.T, raw string) value.Content {
	t.Helper()

	content, err := value.NewContent(raw)
	require.NoError(t, err)

	return content
}
