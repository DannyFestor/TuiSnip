package tagpane_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestPane_View(t *testing.T) {
	t.Parallel()

	t.Run("says there are no Tags yet", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

		assert.Equal(t, []string{"No Tags yet.    "}, viewLines(pane))
	})

	t.Run("cuts the text to a narrow box", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: 6, Height: boxHeight})

		assert.Equal(t, []string{"No Ta…"}, viewLines(pane))
	})

	t.Run("lists each Tag with its Snippet count, the cursor on the first", func(t *testing.T) {
		t.Parallel()

		pane := samplePane(t, newSampleTags(t))

		assert.Equal(t, []string{
			"# DOCKER       3",
			"# go           4",
			"# unused       0",
		}, viewLines(pane))
	})

	t.Run("scrolls to keep the cursor in a short box", func(t *testing.T) {
		t.Parallel()

		pane := paneIn(t, look.Size{Width: boxWidth, Height: 2}).WithTags(newSampleTags(t).counts)

		pane, _ = pressed(pane, keypress.Letter('G'))

		assert.Equal(t, []string{"# go           4", "# UNUSED       0"}, viewLines(pane))
	})
}

func TestPane_Update(t *testing.T) {
	t.Parallel()

	t.Run("moving the cursor selects the Tag under it", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)

		pane, outcomes := pressed(samplePane(t, sample), keypress.Letter('j'))

		assert.Equal(t, []outcome.Outcome{outcome.TagSelected{ID: sample.golang.ID()}}, outcomes)
		assert.Equal(t, selected(sample.golang.ID()), selectedOf(pane))
	})

	t.Run("staying on the same Tag selects nothing", func(t *testing.T) {
		t.Parallel()

		_, outcomes := pressed(samplePane(t, newSampleTags(t)), keypress.Letter('k'))

		assert.Empty(t, outcomes)
	})

	t.Run("moving with no Tags selects nothing", func(t *testing.T) {
		t.Parallel()

		pane, outcomes := pressed(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), keypress.Letter('j'))

		assert.Empty(t, outcomes)
		assert.Equal(t, selected(domain.TagID{}), selectedOf(pane))
	})

	t.Run("ignores keys that move nothing", func(t *testing.T) {
		t.Parallel()

		_, outcomes := pressed(samplePane(t, newSampleTags(t)), keypress.Letter('x'), keypress.Special(tea.KeyEnter))

		assert.Empty(t, outcomes)
	})

	t.Run("draws in the new Styles", func(t *testing.T) {
		t.Parallel()

		box := look.Size{Width: boxWidth, Height: boxHeight}
		light := look.NewStyles(look.SchemeLight)

		restyled, _, _ := paneIn(t, box).Update(look.Restyled{Styles: light})

		assert.Equal(t, styledPaneIn(t, light, box).View(light.Focused), restyled.View(light.Focused))
	})

	t.Run("keeps the Tags and the cursor when restyled", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		box := look.Size{Width: boxWidth, Height: boxHeight}
		light := look.NewStyles(look.SchemeLight)
		moved, _ := pressed(samplePane(t, sample), keypress.Letter('j'))

		restyled, _, _ := moved.Update(look.Restyled{Styles: light})

		styledFromStart, _ := pressed(styledPaneIn(t, light, box).WithTags(sample.counts), keypress.Letter('j'))
		assert.Equal(t, styledFromStart.View(light.Focused), restyled.View(light.Focused))
	})
}

func TestPane_WithTags(t *testing.T) {
	t.Parallel()

	t.Run("keeps the cursor on the same Tag when the list changes", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(samplePane(t, sample), keypress.Letter('j'))
		added := testkit.Tag(t, testkit.TagSpec{ID: testkit.NewSequentialIDs().NewTagID(), Name: "awk"})

		pane = pane.WithTags(append([]browse.TagCount{{Tag: added, SnippetCount: 1}}, sample.counts...))

		assert.Equal(t, selected(sample.golang.ID()), selectedOf(pane))
	})

	t.Run("follows the cursor's Tag to the top of the list", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(samplePane(t, sample), keypress.Letter('j'))

		pane = pane.WithTags(sample.counts[1:])

		assert.Equal(t, selected(sample.golang.ID()), selectedOf(pane))
	})

	t.Run("keeps the cursor's place when its Tag is gone", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(samplePane(t, sample), keypress.Letter('G'))

		pane = pane.WithTags(sample.counts[:2])

		assert.Equal(t, selected(sample.golang.ID()), selectedOf(pane))
	})
}

func TestPane_SelectedName(t *testing.T) {
	t.Parallel()

	sample := newSampleTags(t)
	pane, _ := pressed(samplePane(t, sample), keypress.Letter('j'))

	assert.Equal(t, "go", pane.SelectedName())
	assert.Empty(t, paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}).SelectedName())
}

func TestPane_SelectedTag(t *testing.T) {
	t.Parallel()

	sample := newSampleTags(t)
	pane, _ := pressed(samplePane(t, sample), keypress.Letter('j'))

	tag, ok := pane.SelectedTag()

	assert.True(t, ok)
	assert.Equal(t, sample.golang, tag)

	_, ok = paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}).SelectedTag()

	assert.False(t, ok)
}

func TestPane_ShortHelp(t *testing.T) {
	t.Parallel()

	pane := paneIn(t, look.Size{Width: boxWidth, Height: boxHeight})

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeTags).ShortHelp(), pane.ShortHelp())
	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeTags).FullHelp(), pane.FullHelp())
}

type selection struct {
	id domain.TagID
	ok bool
}

func selected(id domain.TagID) selection {
	return selection{id: id, ok: !id.IsNil()}
}

func selectedOf(pane tagpane.Pane) selection {
	id, ok := pane.Selected()

	return selection{id: id, ok: ok}
}
