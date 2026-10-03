//go:build e2e

package e2e_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/test/keypress"
)

const (
	screenWidth  = 120
	screenHeight = 40
	waitTimeout  = 5 * time.Second
	pollInterval = 10 * time.Millisecond
)

type session struct {
	t       *testing.T
	program *teatest.TestModel
	frame   *lastFrame
}

func open(t *testing.T, app *bootstrap.App) *session {
	t.Helper()

	frame := &lastFrame{mu: sync.Mutex{}, text: ""}
	program := teatest.NewTestModel(
		t,
		framedModel{inner: app.TUI(), frame: frame},
		teatest.WithInitialTermSize(screenWidth, screenHeight),
	)
	t.Cleanup(func() {
		program.Send(keypress.Ctrl('c'))
		program.WaitFinished(t, teatest.WithFinalTimeout(waitTimeout))
	})

	return &session{t: t, program: program, frame: frame}
}

func (s *session) waitForFrame(text string) {
	s.t.Helper()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	deadline := time.After(waitTimeout)

	for !strings.Contains(s.frame.get(), text) {
		select {
		case <-ticker.C:
		case <-deadline:
			s.t.Fatalf("%q never appeared; last frame:\n%s", text, s.frame.get())
		}
	}
}

func (s *session) press(keys ...tea.KeyPressMsg) {
	for _, pressed := range keys {
		s.program.Send(pressed)
	}
}

type framedModel struct {
	inner tea.Model
	frame *lastFrame
}

func (m framedModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m framedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.inner.Update(msg)
	m.inner = next

	return m, cmd
}

func (m framedModel) View() tea.View {
	view := m.inner.View()
	m.frame.set(ansi.Strip(view.Content))

	return view
}

type lastFrame struct {
	mu   sync.Mutex
	text string
}

func (f *lastFrame) set(text string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.text = text
}

func (f *lastFrame) get() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.text
}
