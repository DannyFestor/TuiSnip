package tagchoice_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestChosen_Stored(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Docker"})

	chosen := tagchoice.Of([]domain.Tag{shell, docker})

	assert.Equal(t, []domain.Tag{docker, shell}, chosen.Stored())
	assert.Empty(t, chosen.Created())
}

func TestChosen_Toggled(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})

	t.Run("adds a Tag it does not carry", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of([]domain.Tag{shell}).Toggled(docker)

		assert.Equal(t, []domain.Tag{docker, shell}, chosen.Stored())
		assert.True(t, chosen.Carries(docker))
	})

	t.Run("removes a Tag it carries", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of([]domain.Tag{shell, docker}).Toggled(docker)

		assert.Equal(t, []domain.Tag{shell}, chosen.Stored())
		assert.False(t, chosen.Carries(docker))
	})
}

func TestChosen_ToggledNew(t *testing.T) {
	t.Parallel()

	t.Run("adds a new Tag name, listing new names ignoring case", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of(nil).ToggledNew(tagName(t, "yaml")).ToggledNew(tagName(t, "API"))

		assert.Equal(t, []value.TagName{tagName(t, "API"), tagName(t, "yaml")}, chosen.Created())
		assert.True(t, chosen.CarriesNew(tagName(t, "yaml")))
	})

	t.Run("removes a new Tag name it carries", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of(nil).ToggledNew(tagName(t, "api")).ToggledNew(tagName(t, "api"))

		assert.Empty(t, chosen.Created())
		assert.False(t, chosen.CarriesNew(tagName(t, "api")))
	})
}

func TestChosen_Names(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "Docker"})

	chosen := tagchoice.Of([]domain.Tag{shell, docker}).ToggledNew(tagName(t, "oneliner"))

	assert.Equal(t, []string{"Docker", "oneliner", "shell"}, chosen.Names())
}

func TestChosen_Equal(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	shell := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "shell"})
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})
	original := tagchoice.Of([]domain.Tag{shell})

	tests := []struct {
		name   string
		chosen tagchoice.Chosen
		want   bool
	}{
		{name: "equals itself", chosen: original, want: true},
		{name: "equals a Tag toggled and back", chosen: original.Toggled(docker).Toggled(docker), want: true},
		{name: "differs by a stored Tag", chosen: original.Toggled(docker), want: false},
		{name: "differs by a new Tag", chosen: original.ToggledNew(tagName(t, "api")), want: false},
		{
			name:   "equals a new Tag created and dropped",
			chosen: original.ToggledNew(tagName(t, "api")).ToggledNew(tagName(t, "api")),
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.chosen.Equal(original))
		})
	}
}

func tagName(t *testing.T, raw string) value.TagName {
	t.Helper()

	name, err := value.NewTagName(raw)
	require.NoError(t, err)

	return name
}
