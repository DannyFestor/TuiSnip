package tui_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui"
)

const programTimeout = 5 * time.Second

func TestModelRunsInProgram(t *testing.T) {
	t.Parallel()

	program := teatest.NewTestModel(
		t,
		newModel(t, listerOf(t, sampleSnippets(t)...), NewMockSnippetCopier(t)),
		teatest.WithInitialTermSize(wideWidth, wideHeight),
	)

	teatest.WaitFor(t, program.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("Prune everything"))
	}, teatest.WithDuration(programTimeout))
	program.Send(ctrl('c'))

	assert.IsType(t, tui.Model{}, program.FinalModel(t, teatest.WithFinalTimeout(programTimeout)))
}
