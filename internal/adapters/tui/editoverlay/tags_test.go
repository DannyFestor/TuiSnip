package editoverlay_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	tagEditorFilter = "filter or new Tag:"
	tagEditorHints  = "down move · enter toggle / create · esc close"
)

func TestSession_tags(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	docker := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"})
	oneliner := testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "oneliner"})
	listed := editoverlay.Options{
		Languages: nil,
		Tags:      []browse.TagCount{{Tag: docker, SnippetCount: 2}, {Tag: oneliner, SnippetCount: 1}},
	}

	t.Run("shows the Tags a stored Snippet carries and names the keys that edit them", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi").Retagged([]domain.Tag{oneliner, docker})
		screen := editingStoredWithOptions(t, listed, stored)

		assert.Contains(t, screen.Screen(), "  Tags        #docker #oneliner   (enter or ctrl+t to edit)")
	})

	t.Run("starts a new Snippet with the Destination's Tags and nothing unsaved", func(t *testing.T) {
		t.Parallel()

		filedIn := destination()
		filedIn.Selection = browseselection.WithTag(docker.ID())
		filedIn.Tags = []domain.Tag{docker}
		screen := editingWithOptions(t, listed, filedIn)

		assert.Contains(t, screen.Screen(), "  Tags        #docker   (enter or ctrl+t to edit)")
		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("names only the bound edit key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.OpenField] = []string{}
		screen := editingWith(t, keys)

		assert.Contains(t, screen.Screen(), "(ctrl+t to edit)")
	})

	t.Run("enter on Tags opens the Tag editor", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(toTags()...)
		screen.Press(keypress.Special(tea.KeyEnter))

		assert.Contains(t, screen.Screen(), tagEditorFilter)
		assert.Regexp(t, `docker +2`, screen.Screen())
		assert.Equal(t, tagEditorHints, screen.Hints())
	})

	t.Run("down on Tags moves on to Language", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(toLanguage()...)

		assert.Contains(t, screen.Screen(), "› Language")
		assert.NotContains(t, screen.Screen(), tagEditorFilter)
	})

	t.Run("edit_tags opens the Tag editor from Title", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(editTags())

		assert.Contains(t, screen.Screen(), tagEditorFilter)
	})

	t.Run("edit_tags opens the Tag editor from inside Content", func(t *testing.T) {
		t.Parallel()

		screen := editing(t)

		screen.Press(enterContent()...)
		screen.Press(editTags())

		assert.Contains(t, screen.Screen(), tagEditorFilter)
	})

	t.Run("opens the Tag editor with the configured key", func(t *testing.T) {
		t.Parallel()

		keys := testsettings.Default(t).Keys
		keys[binding.ScopeEditor][binding.EditTags] = []string{"ctrl+g"}
		screen := editingWith(t, keys)

		screen.Press(keypress.Ctrl('g'))

		assert.Contains(t, screen.Screen(), tagEditorFilter)
	})

	t.Run("toggled and created Tags are unsaved changes that the save carries", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(editTags(), keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Typed("api")...)
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), unsavedTitle)
		assert.Contains(t, screen.Screen(), "Tags        #api #docker")

		screen.Press(save())

		want := input("", "", "")
		want.Tags = []domain.Tag{docker}
		want.NewTags = []string{"api"}
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: want}}, screen.Outcomes())
	})

	t.Run("a Tag toggled and toggled back leaves nothing unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(editTags(), keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Special(tea.KeyEscape))

		assert.NotContains(t, screen.Screen(), unsavedTitle)
	})

	t.Run("cancelling after creating a Tag asks first, then discards it unsaved", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(editTags())
		screen.Press(keypress.Typed("api")...)
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape), keypress.Special(tea.KeyEscape))

		assert.Contains(t, screen.Screen(), discardQuestion)

		screen.Press(keypress.Letter('y'))

		assert.False(t, screen.IsOpen())
		assert.Empty(t, screen.Outcomes())
	})

	t.Run("an update carries the Tags left on and the new ones", func(t *testing.T) {
		t.Parallel()

		stored := storedSnippet(t, "echo hi").Retagged([]domain.Tag{docker})
		screen := editingStoredWithOptions(t, listed, stored)

		screen.Press(editTags(), keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyDown))
		screen.Press(keypress.Special(tea.KeyEnter))
		screen.Press(keypress.Typed("yaml")...)
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape), save())

		want := updateInput(stored, "Prune", "echo hi")
		want.Tags = []domain.Tag{oneliner}
		want.NewTags = []string{"yaml"}
		assert.Equal(t, []outcome.Outcome{outcome.UpdateRequested{Input: want}}, screen.Outcomes())
	})

	t.Run("offers the stored Tag for its name typed in another case", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(editTags())
		screen.Press(keypress.Typed("DOCKER")...)
		screen.Press(keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape), save())

		want := input("", "", "")
		want.Tags = []domain.Tag{docker}
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: want}}, screen.Outcomes())
	})

	t.Run("keeps a Language picked after the Tags", func(t *testing.T) {
		t.Parallel()

		screen := editingWithOptions(t, listed, destination())

		screen.Press(editTags(), keypress.Special(tea.KeyEnter), keypress.Special(tea.KeyEscape))
		screen.Press(pickLanguage())
		screen.Press(keypress.Typed("bash")...)
		screen.Press(keypress.Special(tea.KeyEnter), save())

		want := inputIn("", "", "Bash", "")
		want.Tags = []domain.Tag{docker}
		assert.Equal(t, []outcome.Outcome{outcome.SaveRequested{Input: want}}, screen.Outcomes())
	})
}
