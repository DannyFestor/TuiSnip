package snippetlist_test

import (
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 30
	boxHeight = 3
)

func upperCursor() look.FrameStyle {
	return look.FrameStyle{
		Border: lipgloss.NewStyle(),
		Title:  lipgloss.NewStyle(),
		Cursor: lipgloss.NewStyle().Transform(strings.ToUpper),
	}
}

func listOf(t *testing.T, snippets []domain.Snippet) snippetlist.List {
	t.Helper()

	list := snippetlist.New(testsettings.Default(t).Keys, look.NewStyles(), snippetlist.Language)
	list, _, _ = list.Update(look.Resized{Box: look.Size{Width: boxWidth, Height: boxHeight}})

	return list.WithSnippets(snippets)
}

func numberedSnippets(t *testing.T, count int) []domain.Snippet {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	snippets := make([]domain.Snippet, 0, count)

	for number := 1; number <= count; number++ {
		snippets = append(snippets, testkit.Snippet(t, testkit.SnippetSpec{
			ID:       ids.NewSnippetID(),
			Title:    "Snippet " + strconv.Itoa(number),
			Fragment: testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: "Go"},
		}))
	}

	return snippets
}

func pressed(t *testing.T, list snippetlist.List, keys ...tea.KeyPressMsg) snippetlist.List {
	t.Helper()

	for _, key := range keys {
		list, _, _ = list.Update(key)
	}

	return list
}

func selectedTitle(t *testing.T, list snippetlist.List) string {
	t.Helper()

	selected, ok := list.Selected()
	if !ok {
		return ""
	}

	return selected.Title().String()
}

func rows(list snippetlist.List) []string {
	return strings.Split(ansi.Strip(list.View(upperCursor())), "\n")
}

func letter(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func special(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}
