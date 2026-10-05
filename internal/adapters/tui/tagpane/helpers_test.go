package tagpane_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

const (
	boxWidth  = 16
	boxHeight = 3
)

type sampleTags struct {
	docker domain.Tag
	golang domain.Tag
	unused domain.Tag
	counts []browse.TagCount
}

func newSampleTags(t *testing.T) sampleTags {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	sample := sampleTags{
		docker: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"}),
		golang: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "go"}),
		unused: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "unused"}),
		counts: nil,
	}
	sample.counts = []browse.TagCount{
		{Tag: sample.docker, SnippetCount: 3},
		{Tag: sample.golang, SnippetCount: 4},
		{Tag: sample.unused, SnippetCount: 0},
	}

	return sample
}

func upperCursor() look.FrameStyle {
	return look.FrameStyle{
		Border: lipgloss.NewStyle(),
		Title:  lipgloss.NewStyle(),
		Cursor: lipgloss.NewStyle().Transform(strings.ToUpper),
	}
}

func paneIn(t *testing.T, box look.Size) tagpane.Pane {
	t.Helper()

	return styledPaneIn(t, look.NewStyles(look.SchemeDark), box)
}

func styledPaneIn(t *testing.T, styles look.Styles, box look.Size) tagpane.Pane {
	t.Helper()

	pane := tagpane.New(testsettings.Default(t).Keys, styles)
	pane, _, _ = pane.Update(look.Resized{Box: box})

	return pane
}

func samplePane(t *testing.T, sample sampleTags) tagpane.Pane {
	t.Helper()

	return paneIn(t, look.Size{Width: boxWidth, Height: boxHeight}).WithTags(sample.counts)
}

func pressed(pane tagpane.Pane, keys ...tea.KeyPressMsg) (tagpane.Pane, []outcome.Outcome) {
	var all []outcome.Outcome

	for _, key := range keys {
		var outcomes []outcome.Outcome

		pane, outcomes, _ = pane.Update(key)
		all = append(all, outcomes...)
	}

	return pane, all
}

func viewLines(pane tagpane.Pane) []string {
	return strings.Split(ansi.Strip(pane.View(upperCursor())), "\n")
}
