package tageditor_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

func TestEditor_View(t *testing.T) {
	t.Parallel()

	tags := newSampleTags(t)

	t.Run("lists every Tag with its count and marks the ones the Snippet carries", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of([]domain.Tag{tags.docker}))

		assert.Regexp(t, `✓ docker +3`, screen.Screen())
		assert.Regexp(t, `  testing +1`, screen.Screen())
		assert.Regexp(t, `  YAML +0`, screen.Screen())
	})

	t.Run("names the Tags the Snippet carries above the filter", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of([]domain.Tag{tags.yaml, tags.docker}).ToggledNew(tagName(t, "oneliner"))
		screen := editing(t, tags.listed(), chosen)

		assert.Regexp(t, `(?s)Tags  docker, oneliner, YAML.*\n.*filter or new Tag:`, screen.Screen())
	})

	t.Run("lists a new Tag the Snippet carries as new", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil).ToggledNew(tagName(t, "oneliner")))

		assert.Regexp(t, `(?s)docker.*\n.*✓ oneliner +new.*\n.*testing`, screen.Screen())
	})

	t.Run("filters the Tags as the name is typed", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("te")...)

		assert.Contains(t, screen.Screen(), "testing")
		assert.NotContains(t, screen.Screen(), "docker")
	})

	t.Run("closes the list with a create row when no Tag has the typed name", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("te")...)

		assert.Regexp(t, `(?s)testing +1.*\n.*\+ create "te" +new`, screen.Screen())
	})

	t.Run("offers the Tag, not a create row, for its name typed in another case", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("yaml")...)

		assert.Contains(t, screen.Screen(), "YAML")
		assert.NotContains(t, screen.Screen(), "create")
	})

	t.Run("offers no create row for a blank name", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("  ")...)

		assert.NotContains(t, screen.Screen(), "create")
	})

	t.Run("names the typed Tag trimmed on the create row", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed(" api ")...)

		assert.Contains(t, screen.Screen(), `+ create "api"`)
	})

	t.Run("scrolls a long list to keep the cursor in the frame below the Tags line", func(t *testing.T) {
		t.Parallel()

		const count = 20

		screen := editingOn(t, look.Size{Width: 60, Height: 20}, numberedTags(t, count), tagchoice.Of(nil))

		for range count - 1 {
			screen.Press(down())
		}

		assert.Contains(t, screen.Screen(), "tag19")
		assert.NotContains(t, screen.Screen(), "tag00")
	})

	t.Run("says when there are no Tags yet", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, nil, tagchoice.Of(nil))

		assert.Contains(t, screen.Screen(), "No Tags yet.")
	})

	t.Run("hints move, toggle / create and close", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		assert.Equal(t, "down move · enter toggle / create · esc close", screen.Hints())
	})
}

func TestEditor_Update(t *testing.T) {
	t.Parallel()

	tags := newSampleTags(t)

	t.Run("toggles the highlighted Tag and stays open", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of([]domain.Tag{tags.docker}))

		screen.Press(enter(), down(), enter())

		assert.True(t, screen.IsOpen())
		assert.Regexp(t, `  docker +3`, screen.Screen())
		assert.Regexp(t, `✓ testing +1`, screen.Screen())
		assert.Empty(t, screen.Outcomes())
	})

	t.Run("reports the toggles when it closes", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of([]domain.Tag{tags.docker}))

		screen.Press(enter(), down(), enter())

		assert.Equal(t, []domain.Tag{tags.testing}, edited(t, screen).Stored())
	})

	t.Run("creates and adds the typed name on the create row", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed(" api ")...)
		screen.Press(down(), enter())

		assert.Regexp(t, `✓ api +new`, screen.Screen())
		assert.NotContains(t, screen.Screen(), "create")
		assert.Equal(t, []value.TagName{tagName(t, "api")}, edited(t, screen).Created())
	})

	t.Run("drops a created Tag toggled off but keeps listing it", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("api")...)
		screen.Press(enter(), enter())

		assert.Regexp(t, `  api +new`, screen.Screen())
		assert.Empty(t, edited(t, screen).Created())
	})

	t.Run("refuses a name with a comma", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Press(keypress.Typed("a,b")...)
		screen.Press(enter())

		assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: "Tag name contains a comma"}}, screen.Outcomes())
		assert.Contains(t, screen.Screen(), `+ create "a,b"`)
		assert.True(t, screen.IsOpen())
	})

	t.Run("reports the Tags unchanged when closed without a toggle", func(t *testing.T) {
		t.Parallel()

		chosen := tagchoice.Of([]domain.Tag{tags.yaml})
		screen := editing(t, tags.listed(), chosen)

		assert.True(t, edited(t, screen).Equal(chosen))
	})

	t.Run("filters on pasted text", func(t *testing.T) {
		t.Parallel()

		screen := editing(t, tags.listed(), tagchoice.Of(nil))

		screen.Send(tea.PasteMsg{Content: "yam"})
		screen.Press(enter())

		assert.Equal(t, []domain.Tag{tags.yaml}, edited(t, screen).Stored())
	})
}
