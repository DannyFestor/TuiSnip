package overlaytest

import (
	"slices"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/test/screencell"
)

const (
	hintSeparator = " · "
	topLeftCorner = "╭"
)

type Driver struct {
	t        testing.TB
	stack    outcome.Stack
	outcomes []outcome.Outcome
	pending  []tea.Cmd
}

func Open(tb testing.TB, screen look.Size, opened outcome.Overlay) *Driver {
	tb.Helper()

	driver := &Driver{t: tb, stack: outcome.NewStack(), outcomes: nil, pending: nil}
	driver.Send(tea.WindowSizeMsg{Width: screen.Width, Height: screen.Height})
	driver.settle(driver.stack.Pushed(opened))

	return driver
}

func OpenWith(tb testing.TB, screen look.Size, opened outcome.Overlay, started tea.Cmd) *Driver {
	tb.Helper()

	driver := Open(tb, screen, opened)
	driver.queue(started)

	return driver
}

func (d *Driver) RunCommands() {
	d.t.Helper()

	for len(d.pending) > 0 {
		d.RunQueuedCommands()
	}
}

func (d *Driver) RunQueuedCommands() {
	d.t.Helper()

	started := d.pending
	d.pending = nil

	for _, msg := range messagesOf(started) {
		if msg != nil {
			d.Send(msg)
		}
	}
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

func (d *Driver) Click(text string) {
	d.t.Helper()

	d.Send(pointer.Clicked{At: d.CellOf(text), Double: false})
}

func (d *Driver) DoubleClick(text string) {
	d.t.Helper()

	d.Send(pointer.Clicked{At: d.CellOf(text), Double: true})
}

func (d *Driver) CellOf(text string) pointer.Point {
	d.t.Helper()

	return screencell.Find(d.t, d.Screen(), text)
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
	return ansi.Strip(d.StyledScreen())
}

func (d *Driver) StyledScreen() string {
	return d.stack.Render()
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

func (d *Driver) settle(stack outcome.Stack, outcomes []outcome.Outcome, cmd tea.Cmd) {
	d.stack = stack
	d.outcomes = append(d.outcomes, outcomes...)
	d.queue(cmd)
}

func (d *Driver) queue(cmd tea.Cmd) {
	if cmd != nil {
		d.pending = append(d.pending, cmd)
	}
}

// Commands run at once, so the timers of many key presses cost one wait.
func messagesOf(cmds []tea.Cmd) []tea.Msg {
	var running sync.WaitGroup

	results := make([][]tea.Msg, len(cmds))
	for index, cmd := range cmds {
		running.Go(func() { results[index] = messagesFrom(cmd) })
	}

	running.Wait()

	return slices.Concat(results...)
}

func messagesFrom(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		return messagesOf(batch)
	}

	return []tea.Msg{msg}
}
