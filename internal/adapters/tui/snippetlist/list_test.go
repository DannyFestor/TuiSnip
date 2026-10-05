package snippetlist_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

func TestList_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []tea.KeyPressMsg
		want string
	}{
		{name: "down moves to the next Snippet", keys: []tea.KeyPressMsg{keypress.Letter('j')}, want: "Snippet 2"},
		{name: "up stops at the first Snippet", keys: []tea.KeyPressMsg{keypress.Letter('k')}, want: "Snippet 1"},
		{name: "G goes to the last Snippet", keys: []tea.KeyPressMsg{keypress.Letter('G')}, want: "Snippet 5"},
		{
			name: "g returns to the first Snippet",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Letter('g')},
			want: "Snippet 1",
		},
		{
			name: "down stops at the last Snippet",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Letter('j')},
			want: "Snippet 5",
		},
		{
			name: "page down moves by the box height",
			keys: []tea.KeyPressMsg{keypress.Special(tea.KeyPgDown)},
			want: "Snippet 4",
		},
		{
			name: "page down stops at the last Snippet",
			keys: []tea.KeyPressMsg{keypress.Special(tea.KeyPgDown), keypress.Special(tea.KeyPgDown)},
			want: "Snippet 5",
		},
		{
			name: "page up moves by the box height",
			keys: []tea.KeyPressMsg{keypress.Letter('G'), keypress.Special(tea.KeyPgUp)},
			want: "Snippet 2",
		},
		{
			name: "page up stops at the first Snippet",
			keys: []tea.KeyPressMsg{keypress.Letter('j'), keypress.Special(tea.KeyPgUp)},
			want: "Snippet 1",
		},
		{
			name: "ignores keys without a Binding",
			keys: []tea.KeyPressMsg{keypress.Letter('j'), keypress.Letter('x')},
			want: "Snippet 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := pressed(t, listOf(t, numberedSnippets(t, 5)), tt.keys...)

			assert.Equal(t, tt.want, selectedTitle(t, list))
		})
	}

	t.Run("asks to Copy the Snippet under the cursor", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 2)
		list := pressed(t, listOf(t, snippets), keypress.Letter('j'))

		_, outcomes, _ := list.Update(keypress.Letter('y'))

		assert.Equal(t, []outcome.Outcome{outcome.CopyRequested{ID: snippets[1].ID()}}, outcomes)
	})

	t.Run("asks for no Copy with no Snippets", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := listOf(t, nil).Update(keypress.Letter('y'))

		assert.Empty(t, outcomes)
	})

	t.Run("asks for no Copy on other keys", func(t *testing.T) {
		t.Parallel()

		_, outcomes, _ := listOf(t, numberedSnippets(t, 2)).Update(keypress.Letter('j'))

		assert.Empty(t, outcomes)
	})

	t.Run("ignores messages it does not handle", func(t *testing.T) {
		t.Parallel()

		list := listOf(t, numberedSnippets(t, 2))

		next, outcomes, cmd := list.Update(tea.FocusMsg{})

		assert.Equal(t, rows(list), rows(next))
		assert.Empty(t, outcomes)
		assert.Nil(t, cmd)
	})

	t.Run("draws in the new Styles", func(t *testing.T) {
		t.Parallel()

		light := look.NewStyles(look.SchemeLight)

		restyled, _, _ := listOf(t, nil).Update(look.Restyled{Styles: light})

		assert.Equal(t, styledListOf(t, light, nil).View(upperCursor()), restyled.View(upperCursor()))
	})
}

func TestList_View(t *testing.T) {
	t.Parallel()

	t.Run("shows the empty hint with no Snippets", func(t *testing.T) {
		t.Parallel()

		got := rows(listOf(t, nil))

		assert.Equal(t, "No Snippets here.", got[0])
		assert.Contains(t, got, "n  new Snippet")
	})

	t.Run("shows each Snippet's title and meta, the cursor row marked", func(t *testing.T) {
		t.Parallel()

		got := rows(listOf(t, numberedSnippets(t, 2)))

		assert.Equal(t, []string{
			"SNIPPET 1                   GO",
			"Snippet 2                   Go",
		}, got)
	})

	t.Run("scrolls down to keep the cursor in view", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, listOf(t, numberedSnippets(t, 5)), keypress.Letter('G'))

		assert.Equal(t, []string{
			"Snippet 3                   Go",
			"Snippet 4                   Go",
			"SNIPPET 5                   GO",
		}, rows(list))
	})

	t.Run("scrolls back up to keep the cursor in view", func(t *testing.T) {
		t.Parallel()

		list := pressed(
			t,
			listOf(t, numberedSnippets(t, 5)),
			keypress.Letter('G'),
			keypress.Letter('k'),
			keypress.Letter('k'),
			keypress.Letter('k'),
		)

		assert.Equal(t, []string{
			"SNIPPET 2                   GO",
			"Snippet 3                   Go",
			"Snippet 4                   Go",
		}, rows(list))
	})

	t.Run("keeps the cursor in view when the box shrinks", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, listOf(t, numberedSnippets(t, 5)), keypress.Letter('j'), keypress.Letter('j'))

		list, _, _ = list.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: 1}})

		assert.Equal(t, []string{"SNIPPET 3                   GO"}, rows(list))
	})

	t.Run("shows the meta the parent chose", func(t *testing.T) {
		t.Parallel()

		list := snippetlist.New(
			testsettings.Default(t).Keys,
			look.NewStyles(look.SchemeDark),
			func(domain.Snippet) string { return "Root" },
		)
		list, _, _ = list.Update(look.Resized{Box: look.Size{Width: 12, Height: 1}})
		list = list.WithSnippets(numberedSnippets(t, 1))

		assert.Equal(t, []string{"SNIPPE… ROOT"}, rows(list))
	})
}

func TestList_WithMeta(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedSnippets(t, 1))
	list, _, _ = list.Update(look.Resized{Box: look.Size{Width: 12, Height: 1}})

	list = list.WithMeta(func(domain.Snippet) string { return "Root" })

	assert.Equal(t, []string{"SNIPPE… ROOT"}, rows(list))
}

func TestList_WithSnippets(t *testing.T) {
	t.Parallel()

	t.Run("keeps the cursor on the last Snippet of a shorter list", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 5)
		list := pressed(t, listOf(t, snippets), keypress.Letter('G'))

		list = list.WithSnippets(snippets[:2])

		assert.Equal(t, "Snippet 2", selectedTitle(t, list))
	})

	t.Run("selects nothing in an empty list", func(t *testing.T) {
		t.Parallel()

		_, ok := listOf(t, nil).Selected()

		assert.False(t, ok)
	})

	t.Run("lists the Snippets it was given", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 2)

		assert.Equal(t, snippets, listOf(t, snippets).Snippets())
	})
}

func TestList_WithCursorOn(t *testing.T) {
	t.Parallel()

	t.Run("moves the cursor to the Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 5)

		list := listOf(t, snippets).WithCursorOn(snippets[3].ID())

		assert.Equal(t, "Snippet 4", selectedTitle(t, list))
	})

	t.Run("moves the cursor back to the first Snippet", func(t *testing.T) {
		t.Parallel()

		snippets := numberedSnippets(t, 5)
		list := pressed(t, listOf(t, snippets), keypress.Letter('G'))

		list = list.WithCursorOn(snippets[0].ID())

		assert.Equal(t, "Snippet 1", selectedTitle(t, list))
	})

	t.Run("keeps the cursor for a Snippet not in the list", func(t *testing.T) {
		t.Parallel()

		list := pressed(t, listOf(t, numberedSnippets(t, 5)), keypress.Letter('j'))
		list = list.WithCursorOn(domain.SnippetID{})

		assert.Equal(t, "Snippet 2", selectedTitle(t, list))
	})
}

func TestList_Moved(t *testing.T) {
	t.Parallel()

	list := listOf(t, numberedSnippets(t, 5)).Moved(move.Bottom).Moved(move.None)

	assert.Equal(t, "Snippet 5", selectedTitle(t, list))
}

func TestList_ShortHelp(t *testing.T) {
	t.Parallel()

	list := listOf(t, nil)

	assert.Equal(t, testsettings.Default(t).Keys.For(binding.ScopeSnippetList).ShortHelp(), list.ShortHelp())
	assert.Equal(t, [][]key.Binding{list.ShortHelp()}, list.FullHelp())
}
