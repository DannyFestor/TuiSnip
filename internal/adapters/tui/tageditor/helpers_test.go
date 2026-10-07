package tageditor_test

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tageditor"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/overlaytest"
	"github.com/DannyFestor/TuiSnip/test/testsettings"
)

type sampleTags struct {
	docker  domain.Tag
	testing domain.Tag
	yaml    domain.Tag
}

func newSampleTags(t *testing.T) sampleTags {
	t.Helper()

	ids := testkit.NewSequentialIDs()

	return sampleTags{
		docker:  testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "docker"}),
		testing: testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "testing"}),
		yaml:    testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: "YAML"}),
	}
}

func (s sampleTags) listed() []browse.TagCount {
	return []browse.TagCount{
		{Tag: s.docker, SnippetCount: 3},
		{Tag: s.testing, SnippetCount: 1},
		{Tag: s.yaml, SnippetCount: 0},
	}
}

func editing(t *testing.T, listed []browse.TagCount, chosen tagchoice.Chosen) *overlaytest.Driver {
	t.Helper()

	return editingOn(t, look.Size{Width: 120, Height: 40}, listed, chosen)
}

func editingOn(
	t *testing.T, screen look.Size, listed []browse.TagCount, chosen tagchoice.Chosen,
) *overlaytest.Driver {
	t.Helper()

	return editingStyled(t, look.NewStyles(look.SchemeDark), screen, tageditor.Offer{Listed: listed, Chosen: chosen})
}

func editingStyled(t *testing.T, styles look.Styles, screen look.Size, offer tageditor.Offer) *overlaytest.Driver {
	t.Helper()

	opened, _ := tageditor.New(testsettings.Default(t).Keys, styles, offer)

	return overlaytest.Open(t, screen, opened)
}

func numberedTags(t *testing.T, count int) []browse.TagCount {
	t.Helper()

	ids := testkit.NewSequentialIDs()
	listed := make([]browse.TagCount, 0, count)

	for number := range count {
		name := fmt.Sprintf("tag%02d", number)
		listed = append(listed, browse.TagCount{
			Tag:          testkit.Tag(t, testkit.TagSpec{ID: ids.NewTagID(), Name: name}),
			SnippetCount: number,
		})
	}

	return listed
}

func edited(t *testing.T, screen *overlaytest.Driver) tagchoice.Chosen {
	t.Helper()

	screen.Press(keypress.Special(tea.KeyEscape))
	require.False(t, screen.IsOpen())
	require.Len(t, screen.Outcomes(), 1)

	reported, ok := screen.Outcomes()[0].(outcome.TagsEdited)
	require.True(t, ok)

	return reported.Chosen
}

func tagName(t *testing.T, raw string) value.TagName {
	t.Helper()

	name, err := value.NewTagName(raw)
	require.NoError(t, err)

	return name
}

func enter() tea.KeyPressMsg {
	return keypress.Special(tea.KeyEnter)
}

func down() tea.KeyPressMsg {
	return keypress.Special(tea.KeyDown)
}
