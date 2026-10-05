package tagpane_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const tallBoxHeight = 5

func TestPane_UpdateNewTag(t *testing.T) {
	t.Parallel()

	t.Run("asks to create the typed Tag", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'))
		pane, _ = pressed(pane, keypress.Typed(" awk ")...)

		pane, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagCreateRequested{Input: tag.CreateInput{Name: " awk "}}}, outcomes)
		assert.False(t, pane.Naming())
	})

	t.Run("creates the first Tag", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), keypress.Letter('N'))
		pane, _ = pressed(pane, keypress.Typed("awk")...)

		_, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagCreateRequested{Input: tag.CreateInput{Name: "awk"}}}, outcomes)
	})

	refusals := []struct {
		name  string
		typed string
		want  string
	}{
		{name: "refuses a name a Tag has, ignoring case", typed: "GO", want: "Tag go already exists"},
		{name: "refuses the name of the first Tag", typed: "docker", want: "Tag docker already exists"},
		{name: "refuses a blank name", typed: "  ", want: "Tag name is blank"},
		{name: "refuses a name with a comma", typed: "go,rust", want: "Tag name contains a comma"},
	}
	for _, tt := range refusals {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'))
			pane, _ = pressed(pane, keypress.Typed(tt.typed)...)

			pane, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

			assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: tt.want}}, outcomes)
			assert.True(t, pane.Naming(), "the row stays open to fix the name")
		})
	}
}

func TestPane_UpdateRename(t *testing.T) {
	t.Parallel()

	t.Run("renames the Tag under the cursor, starting from its name", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(tallPane(t, sample), keypress.Letter('j'), keypress.Letter('r'))
		pane, _ = pressed(pane, keypress.Typed("lang")...)

		pane, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagRenameRequested{
			Input: tag.RenameInput{TagID: sample.golang.ID(), Name: "golang"},
		}}, outcomes)
		assert.False(t, pane.Naming())
	})

	t.Run("asks to rename onto a name another Tag has, which merges them", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(tallPane(t, sample), keypress.Letter('j'), keypress.Letter('r'))
		pane, _ = pressed(pane, keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyBackspace))
		pane, _ = pressed(pane, keypress.Typed("Docker")...)

		_, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagRenameRequested{
			Input: tag.RenameInput{TagID: sample.golang.ID(), Name: "Docker"},
		}}, outcomes)
	})

	t.Run("refuses a blank name", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('j'), keypress.Letter('r'))
		pane, _ = pressed(pane, keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyBackspace))

		_, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.NoticeShown{Text: "Tag name is blank"}}, outcomes)
	})

	t.Run("does nothing without Tags", func(t *testing.T) {
		t.Parallel()

		pane, outcomes := pressed(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), keypress.Letter('r'))

		assert.Empty(t, outcomes)
		assert.False(t, pane.Naming())
	})
}

func TestPane_UpdateDelete(t *testing.T) {
	t.Parallel()

	t.Run("asks to delete the Tag under the cursor", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)

		_, outcomes := pressed(samplePane(t, sample), keypress.Letter('j'), keypress.Letter('d'))

		assert.Equal(t, []outcome.Outcome{
			outcome.TagSelected{ID: sample.golang.ID()},
			outcome.TagDeleteAsked{ID: sample.golang.ID()},
		}, outcomes)
	})

	t.Run("does nothing without Tags", func(t *testing.T) {
		t.Parallel()

		_, outcomes := pressed(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), keypress.Letter('d'))

		assert.Empty(t, outcomes)
	})
}

func TestPane_UpdateWhileNaming(t *testing.T) {
	t.Parallel()

	t.Run("cancel drops the typed row", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'), keypress.Letter('x'))

		pane, outcomes := pressed(pane, keypress.Special(tea.KeyEscape))

		assert.Empty(t, outcomes)
		assert.False(t, pane.Naming())
		assert.Len(t, viewLines(pane), 3)
	})

	t.Run("movement keys type instead of moving the cursor", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(tallPane(t, sample), keypress.Letter('N'))

		pane, outcomes := pressed(pane, keypress.Letter('j'))

		assert.Empty(t, outcomes)
		assert.Equal(t, selected(sample.docker.ID()), selectedOf(pane))
		assert.True(t, pane.Naming())
	})

	t.Run("takes a paste", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'))
		pane, _, _ = pane.Update(tea.PasteMsg{Content: "awk"})

		_, outcomes := pressed(pane, keypress.Special(tea.KeyEnter))

		assert.Equal(t, []outcome.Outcome{outcome.TagCreateRequested{Input: tag.CreateInput{Name: "awk"}}}, outcomes)
	})
}

func TestPane_ViewWhileNaming(t *testing.T) {
	t.Parallel()

	t.Run("types a new Tag on the row where its name sorts", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'))
		pane, _ = pressed(pane, keypress.Typed("hx")...)

		lines := viewLines(pane)

		require.Len(t, lines, 4)
		assert.Equal(t, "# go           4", lines[1])
		assert.Regexp(t, `^# HX +0$`, lines[2])
		assert.Equal(t, "# unused       0", lines[3])
	})

	sorted := []struct {
		name      string
		typed     string
		wantIndex int
	}{
		{name: "an empty name sorts first", typed: "", wantIndex: 0},
		{name: "a name before every Tag sorts first", typed: "awk", wantIndex: 0},
		{name: "a name after every Tag sorts last", typed: "zsh", wantIndex: 3},
		{name: "sorts ignoring case", typed: "Hx", wantIndex: 2},
		{name: "sorts by the trimmed name", typed: "  zsh", wantIndex: 3},
		{name: "sorts a name the next Tag starts with before it", typed: "unuse", wantIndex: 2},
		{name: "sorts a name with a comma as typed", typed: "zsh,", wantIndex: 3},
	}
	for _, tt := range sorted {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('j'), keypress.Letter('N'))
			pane, _ = pressed(pane, keypress.Typed(tt.typed)...)

			lines := viewLines(pane)

			require.Len(t, lines, 4)
			assert.Regexp(t, `0$`, lines[tt.wantIndex])
			assert.Equal(t, strings.ToUpper(lines[tt.wantIndex]), lines[tt.wantIndex], "the cursor is on the typed row")
		})
	}

	t.Run("moves the row as the name is typed", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'), keypress.Letter('z'))
		pane, _ = pressed(pane, keypress.Special(tea.KeyBackspace), keypress.Letter('a'))

		lines := viewLines(pane)

		require.Len(t, lines, 4)
		assert.Regexp(t, `^# A +0$`, lines[0])
		assert.Equal(t, "# docker       3", lines[1])
	})

	t.Run("types the first Tag in place of the empty text", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}), keypress.Letter('N'))

		assert.Regexp(t, `^# +0$`, viewLines(pane)[0])
	})

	t.Run("keeps the renamed row when the Tags reload empty", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('j'), keypress.Letter('r'))

		pane = pane.WithTags(nil)

		assert.Regexp(t, `^# GO +4$`, viewLines(pane)[0])
	})

	t.Run("types a rename in place of the row", func(t *testing.T) {
		t.Parallel()

		pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('j'), keypress.Letter('r'))

		lines := viewLines(pane)

		require.Len(t, lines, 3)
		assert.Regexp(t, `^# GO +4$`, lines[1])
	})
}

func TestPane_ShortHelpWhileNaming(t *testing.T) {
	t.Parallel()

	pane, _ := pressed(tallPane(t, newSampleTags(t)), keypress.Letter('N'))

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeNameInput).ShortHelp(), pane.ShortHelp())
}

func TestPane_WithCursorOn(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor onto the Tag", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)

		pane := tallPane(t, sample).WithCursorOn(sample.unused.ID())

		assert.Equal(t, selected(sample.unused.ID()), selectedOf(pane))
	})

	t.Run("moves the cursor onto the first Tag", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(tallPane(t, sample), keypress.Letter('G'))

		pane = pane.WithCursorOn(sample.docker.ID())

		assert.Equal(t, selected(sample.docker.ID()), selectedOf(pane))
	})

	t.Run("keeps the cursor for a Tag it does not list", func(t *testing.T) {
		t.Parallel()

		sample := newSampleTags(t)
		pane, _ := pressed(tallPane(t, sample), keypress.Letter('j'))

		ids := testkit.NewSequentialIDs()
		for range sample.counts {
			ids.NewTagID()
		}

		pane = pane.WithCursorOn(ids.NewTagID())

		assert.Equal(t, selected(sample.golang.ID()), selectedOf(pane))
	})
}

func tallPane(t *testing.T, sample sampleTags) tagpane.Pane {
	t.Helper()

	return paneIn(t, look.Size{Width: boxWidth, Height: tallBoxHeight}).WithTags(sample.counts)
}
