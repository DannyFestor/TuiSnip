//go:build e2e

package e2e_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/test/keypress"
	"github.com/DannyFestor/TuiSnip/test/testapp"
)

func TestCreateMergeAndDeleteTags(t *testing.T) {
	t.Parallel()

	_, app := testapp.Start(t, testapp.RecordingTool)
	golang := testapp.SeedTag(t, app, tag.CreateInput{Name: "go"})
	merged := testapp.SeedTag(t, app, tag.CreateInput{Name: "golang"})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Client", Tags: []domain.Tag{golang, merged}})
	testapp.SeedSnippet(t, app, snippet.CreateInput{Title: "Pods", Tags: []domain.Tag{merged}})
	screen := open(t, app)
	screen.waitForFrame("# golang")

	screen.press(keypress.Letter('2'), keypress.Letter('N'))
	screen.press(keypress.Typed("awk")...)
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("3 # awk · by title")

	screen.press(keypress.Letter('G'), keypress.Letter('r'))
	screen.press(keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyBackspace))
	screen.press(keypress.Special(tea.KeyBackspace), keypress.Special(tea.KeyBackspace))
	screen.press(keypress.Special(tea.KeyEnter))
	screen.waitForFrame("3 # go · by title")
	screen.waitForFrame("Root · plaintext · #go    ")
	assert.NotContains(t, screen.frame.get(), "golang")

	screen.press(keypress.Letter('d'))
	screen.waitForFrame("2 Snippets")
	screen.press(keypress.Letter('y'))
	screen.waitForFrame("3 # awk · by title")

	screen.press(keypress.Letter('d'))
	screen.waitForFrame("No Tags yet.")
	screen.waitForFrame("3 Root · by title")
}
