package overlaytest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

const (
	hintSeparator = " · "
	topLeftCorner = "╭"
)

type Driver struct {
	t        *testing.T
	stack    outcome.Stack
	outcomes []outcome.Outcome
}

func Open(t *testing.T, screen look.Size, opened outcome.Overlay) *Driver {
	t.Helper()

	driver := &Driver{t: t, stack: outcome.NewStack(), outcomes: nil}
	driver.Send(tea.WindowSizeMsg{Width: screen.Width, Height: screen.Height})
	driver.settle(driver.stack.Pushed(opened))

	return driver
}

func (d *Driver) Send(msg tea.Msg) {
	d.t.Helper()

	d.settle(d.stack.Update(msg))
}

func (d *Driver) Press(keys ...tea.KeyPressMsg) {
	d.t.Helper()

	for _, pressed := range keys {
		d.Send(pressed)
	}
}

func (d *Driver) Offer(offered outcome.Outcome) {
	d.t.Helper()

	d.settle(d.stack.Offered(offered))
}

func (d *Driver) Outcomes() []outcome.Outcome {
	return d.outcomes
}

func (d *Driver) IsOpen() bool {
	return d.stack.Open()
}

func (d *Driver) Screen() string {
	return ansi.Strip(d.stack.Render())
}

func (d *Driver) TopBorderWidth() int {
	for line := range strings.SplitSeq(d.Screen(), "\n") {
		if strings.Contains(line, topLeftCorner) {
			return ansi.StringWidth(strings.TrimSpace(line))
		}
	}

	return 0
}

func (d *Driver) Hints() string {
	entries := make([]string, 0, len(d.stack.ShortHelp()))

	for _, hint := range d.stack.ShortHelp() {
		if hint.Enabled() {
			entries = append(entries, hint.Help().Key+" "+hint.Help().Desc)
		}
	}

	return strings.Join(entries, hintSeparator)
}

func (d *Driver) settle(stack outcome.Stack, outcomes []outcome.Outcome, _ tea.Cmd) {
	d.stack = stack
	d.outcomes = append(d.outcomes, outcomes...)
}
