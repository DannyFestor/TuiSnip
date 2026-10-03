package snippetpane_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth    = 24
	codeLines   = 3
	headerLines = 4
	lineCount   = 10
)

func created() time.Time {
	return time.Date(2026, time.September, 6, 9, 0, 0, 0, time.UTC)
}

func paneIn(t *testing.T, location *time.Location, box look.Size) snippetpane.Pane {
	t.Helper()

	pane := snippetpane.New(testsettings.Default(t).Keys, look.NewStyles(), location)
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}

func showing(t *testing.T, snippet domain.Snippet) snippetpane.Pane {
	t.Helper()

	return paneIn(t, time.UTC, look.Size{Width: boxWidth, Height: headerLines + codeLines}).Showing(snippet)
}

func longSnippet(t *testing.T) domain.Snippet {
	t.Helper()

	return snippetWith(t, testkit.SnippetSpec{})
}

func snippetWith(t *testing.T, spec testkit.SnippetSpec) domain.Snippet {
	t.Helper()

	lines := make([]string, 0, lineCount)
	for number := 1; number <= lineCount; number++ {
		lines = append(lines, "line "+strconv.Itoa(number))
	}

	ids := testkit.NewSequentialIDs()
	spec.ID = ids.NewSnippetID()
	spec.Title = "Long"
	spec.Fragment = testkit.FragmentSpec{ID: ids.NewFragmentID(), Language: "Go", Content: strings.Join(lines, "\n")}
	spec.CreatedAt = created()

	return testkit.Snippet(t, spec)
}

func pressed(t *testing.T, pane snippetpane.Pane, keys ...tea.KeyPressMsg) snippetpane.Pane {
	t.Helper()

	for _, key := range keys {
		pane, _, _ = pane.Update(key)
	}

	return pane
}

func lines(pane snippetpane.Pane) []string {
	return strings.Split(ansi.Strip(pane.View()), "\n")
}

func code(pane snippetpane.Pane) []string {
	return lines(pane)[headerLines:]
}

func trimmed(lines []string) []string {
	trimmedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmedLines = append(trimmedLines, strings.TrimRight(line, " "))
	}

	return trimmedLines
}
