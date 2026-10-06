package domain_test

import (
	"math"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const folderUUID = "0192f0c1-7a3b-7c4d-8e5f-00000000f01d"

func TestSnippet_New(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	fragment := testkit.Fragment(t, testkit.FragmentSpec{})

	tests := []struct {
		name      string
		id        domain.SnippetID
		fragments []domain.Fragment
		createdAt time.Time
		updatedAt time.Time
		wantErrs  []error
	}{
		{
			name:      "accepts one fragment",
			id:        snippetID(),
			fragments: []domain.Fragment{fragment},
			createdAt: created,
			updatedAt: created,
		},
		{
			name:      "rejects the nil id",
			id:        domain.SnippetID{},
			fragments: []domain.Fragment{fragment},
			createdAt: created,
			updatedAt: created,
			wantErrs:  []error{domain.ErrNilID},
		},
		{
			name:      "rejects no fragment",
			id:        snippetID(),
			fragments: nil,
			createdAt: created,
			updatedAt: created,
			wantErrs:  []error{domain.ErrNotOneFragment},
		},
		{
			name:      "rejects two fragments",
			id:        snippetID(),
			fragments: []domain.Fragment{fragment, fragment},
			createdAt: created,
			updatedAt: created,
			wantErrs:  []error{domain.ErrNotOneFragment},
		},
		{
			name:      "rejects the epoch",
			id:        snippetID(),
			fragments: []domain.Fragment{fragment},
			createdAt: time.Unix(0, 0),
			updatedAt: created,
			wantErrs:  []error{domain.ErrTimestampOutOfRange},
		},
		{
			name:      "rejects a time past int64 nanoseconds",
			id:        snippetID(),
			fragments: []domain.Fragment{fragment},
			createdAt: created,
			updatedAt: time.Unix(0, math.MaxInt64).Add(time.Nanosecond),
			wantErrs:  []error{domain.ErrTimestampOutOfRange},
		},
		{
			name:      "rejects an update before creation",
			id:        snippetID(),
			fragments: []domain.Fragment{fragment},
			createdAt: created,
			updatedAt: created.Add(-time.Second),
			wantErrs:  []error{domain.ErrUpdatedBeforeCreate},
		},
		{
			name:      "reports every broken rule",
			id:        domain.SnippetID{},
			fragments: nil,
			createdAt: created,
			updatedAt: created.Add(-time.Second),
			wantErrs:  []error{domain.ErrNilID, domain.ErrNotOneFragment, domain.ErrUpdatedBeforeCreate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewSnippet(
				tt.id,
				mustTitle(t, "curl json"),
				value.Description{},
				domain.FolderID{},
				tt.fragments,
				nil,
				tt.createdAt,
				tt.updatedAt,
			)

			requireErrors(t, err, tt.wantErrs)
		})
	}
}

func TestSnippet_Accessors(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	title := mustTitle(t, "curl json")
	description := mustDescription(t, "POST with a JSON body")
	folderID := domain.FolderID(uuid.MustParse(folderUUID))
	fragment := testkit.Fragment(t, testkit.FragmentSpec{Content: "curl -d @body.json"})
	tag := testkit.Tag(t, testkit.TagSpec{Name: "http"})

	snippet, err := domain.NewSnippet(
		snippetID(), title, description, folderID, []domain.Fragment{fragment}, []domain.Tag{tag}, created, updated,
	)

	require.NoError(t, err)
	assert.Equal(t, snippetID(), snippet.ID())
	assert.Equal(t, title, snippet.Title())
	assert.Equal(t, description, snippet.Description())
	assert.Equal(t, folderID, snippet.FolderID())
	assert.Equal(t, []domain.Fragment{fragment}, snippet.Fragments())
	assert.Equal(t, []domain.Tag{tag}, snippet.Tags())
	assert.Equal(t, created, snippet.CreatedAt())
	assert.Equal(t, updated, snippet.UpdatedAt())
}

func TestSnippet_AtRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		folderID domain.FolderID
		want     bool
	}{
		{name: "zero folder id is the Root", folderID: domain.FolderID{}, want: true},
		{name: "a folder id is not the Root", folderID: domain.FolderID(uuid.MustParse(folderUUID)), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snippet := testkit.Snippet(t, testkit.SnippetSpec{FolderID: tt.folderID})

			assert.Equal(t, tt.want, snippet.AtRoot())
		})
	}
}

func TestSnippet_Fragments(t *testing.T) {
	t.Parallel()

	fragments := []domain.Fragment{testkit.Fragment(t, testkit.FragmentSpec{Content: "original"})}
	snippet, err := domain.NewSnippet(
		snippetID(), mustTitle(t, "curl json"), value.Description{}, domain.FolderID{}, fragments, nil,
		time.Unix(1, 0), time.Unix(1, 0),
	)
	require.NoError(t, err)

	fragments[0] = testkit.Fragment(t, testkit.FragmentSpec{Content: "changed by caller"})
	snippet.Fragments()[0] = testkit.Fragment(t, testkit.FragmentSpec{Content: "changed through accessor"})

	assert.Equal(t, "original", snippet.Fragments()[0].Content().String())
}

func TestSnippet_Tags(t *testing.T) {
	t.Parallel()

	t.Run("lists the Tags by name ignoring case", func(t *testing.T) {
		t.Parallel()

		ids := testkit.NewSequentialIDs()
		shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
		docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Docker"})

		snippet := testkit.Snippet(t, testkit.SnippetSpec{Tags: []domain.Tag{shell, docker}})

		assert.Equal(t, []domain.Tag{docker, shell}, snippet.Tags())
	})

	t.Run("keeps its Tags from the caller and the accessor", func(t *testing.T) {
		t.Parallel()

		tags := []domain.Tag{testkit.Tag(t, testkit.TagSpec{Name: "original"})}
		snippet := testkit.Snippet(t, testkit.SnippetSpec{Tags: tags})

		tags[0] = testkit.Tag(t, testkit.TagSpec{Name: "changed by caller"})
		snippet.Tags()[0] = testkit.Tag(t, testkit.TagSpec{Name: "changed through accessor"})

		assert.Equal(t, "original", snippet.Tags()[0].Name().String())
	})
}

func TestSnippet_Retagged(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	http := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "http"})
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Docker"})
	stored := testkit.Snippet(t, testkit.SnippetSpec{Tags: []domain.Tag{http}})

	retagged := stored.Retagged([]domain.Tag{shell, docker})

	assert.Equal(t, []domain.Tag{docker, shell}, retagged.Tags(), "listed by name ignoring case")
	assert.Equal(t, []domain.Tag{http}, stored.Tags(), "the original is untouched")
	assert.Equal(t, stored.UpdatedAt(), retagged.UpdatedAt())
}

func TestSnippet_FirstFragment(t *testing.T) {
	t.Parallel()

	snippet := testkit.Snippet(t, testkit.SnippetSpec{Fragment: testkit.FragmentSpec{Content: "echo hi"}})

	assert.Equal(t, "echo hi", snippet.FirstFragment().Content().String())
}

func TestSnippet_Edit(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	later := created.Add(time.Hour)
	stored := testkit.Snippet(t, testkit.SnippetSpec{
		ID:        snippetID(),
		Title:     "curl",
		FolderID:  domain.FolderID(uuid.MustParse(storedID)),
		Fragment:  testkit.FragmentSpec{Content: "curl", CreatedAt: created},
		Tags:      []domain.Tag{testkit.Tag(t, testkit.TagSpec{Name: "http"})},
		CreatedAt: created,
	})

	t.Run("takes the fields and the time", func(t *testing.T) {
		t.Parallel()

		edited, err := stored.Edit(
			mustTitle(t, "curl json"),
			mustDescription(t, "POST"),
			mustLanguage(t, "Bash"),
			mustContent(t, "curl -d @b.json"),
			later,
		)

		require.NoError(t, err)
		assert.Equal(t, "curl json", edited.Title().String())
		assert.Equal(t, "POST", edited.Description().String())
		assert.Equal(t, "Bash", edited.FirstFragment().Language().String())
		assert.Equal(t, "curl -d @b.json", edited.FirstFragment().Content().String())
		assert.Equal(t, later, edited.UpdatedAt())
		assert.Equal(t, later, edited.FirstFragment().UpdatedAt())
		assert.Equal(t, "curl", stored.Title().String(), "the original is untouched")
	})

	t.Run("keeps the ID, Folder, Fragment, Tags and creation time", func(t *testing.T) {
		t.Parallel()

		edited, err := stored.Edit(
			mustTitle(t, "curl"), value.Description{}, value.PlainText(), mustContent(t, "curl"), later,
		)

		require.NoError(t, err)
		assert.Equal(t, stored.ID(), edited.ID())
		assert.Equal(t, stored.FolderID(), edited.FolderID())
		assert.Equal(t, stored.Tags(), edited.Tags())
		assert.Equal(t, stored.FirstFragment().ID(), edited.FirstFragment().ID())
		assert.Equal(t, created, edited.CreatedAt())
		assert.Equal(t, created, edited.FirstFragment().UpdatedAt(), "unchanged content keeps its time")
	})

	t.Run("rejects a time before creation", func(t *testing.T) {
		t.Parallel()

		_, err := stored.Edit(
			mustTitle(t, "curl json"), value.Description{}, value.PlainText(), mustContent(t, "curl"), created.Add(-1),
		)

		require.ErrorIs(t, err, domain.ErrUpdatedBeforeCreate)
	})
}

func TestSnippet_Duplicate(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	later := created.Add(time.Hour)
	ids := testkit.NewSequentialIDs()
	duplicateID, duplicateFragmentID := ids.NewSnippetID(), ids.NewFragmentID()
	stored := testkit.Snippet(t, testkit.SnippetSpec{
		ID:          snippetID(),
		Title:       "curl",
		Description: "POST",
		FolderID:    domain.FolderID(uuid.MustParse(folderUUID)),
		Fragment:    testkit.FragmentSpec{Language: "Bash", Content: "curl -d @b.json", CreatedAt: created},
		Tags:        []domain.Tag{testkit.Tag(t, testkit.TagSpec{Name: "http"})},
		CreatedAt:   created,
	})

	t.Run("keeps every field but the ids and times", func(t *testing.T) {
		t.Parallel()

		duplicate, err := stored.Duplicate(duplicateID, duplicateFragmentID, later)

		require.NoError(t, err)
		assert.Equal(t, duplicateID, duplicate.ID())
		assert.Equal(t, stored.Title(), duplicate.Title())
		assert.Equal(t, stored.Description(), duplicate.Description())
		assert.Equal(t, stored.FolderID(), duplicate.FolderID())
		assert.Equal(t, stored.Tags(), duplicate.Tags())
		assert.Equal(t, later, duplicate.CreatedAt())
		assert.Equal(t, later, duplicate.UpdatedAt())
		assert.Equal(t, duplicateFragmentID, duplicate.FirstFragment().ID())
		assert.Equal(t, stored.FirstFragment().Language(), duplicate.FirstFragment().Language())
		assert.Equal(t, stored.FirstFragment().Content(), duplicate.FirstFragment().Content())
	})

	t.Run("rejects nil ids", func(t *testing.T) {
		t.Parallel()

		_, err := stored.Duplicate(domain.SnippetID{}, domain.FragmentID{}, later)

		require.ErrorIs(t, err, domain.ErrNilID)
	})
}

func snippetID() domain.SnippetID {
	return domain.SnippetID(uuid.MustParse(storedID))
}

func mustTitle(t *testing.T, raw string) value.Title {
	t.Helper()

	title, err := value.NewTitle(raw)
	require.NoError(t, err)

	return title
}

func mustDescription(t *testing.T, raw string) value.Description {
	t.Helper()

	description, err := value.NewDescription(raw)
	require.NoError(t, err)

	return description
}
