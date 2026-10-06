package tageditor_test

import (
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

	opened, _ := tageditor.New(
		testsettings.Default(t).Keys,
		look.NewStyles(look.SchemeDark),
		tageditor.Offer{Listed: listed, Chosen: chosen},
	)

	return overlaytest.Open(t, look.Size{Width: 120, Height: 40}, opened)
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
