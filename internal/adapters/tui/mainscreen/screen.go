package mainscreen

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
)

const hintWidthDivisor = 2

type Screen struct {
	keys            binding.Keys
	global          binding.Set
	styles          look.Styles
	box             look.Size
	layout          layout
	focus           pane
	selectionHolder pane
	panes           panes
	status          string
}

func New(keys binding.Keys, styles look.Styles, location *time.Location) Screen {
	return Screen{
		keys:            keys,
		global:          keys.For(binding.ScopeGlobal),
		styles:          styles,
		box:             look.Size{Width: 0, Height: 0},
		layout:          arrange(look.Size{Width: 0, Height: 0}, paneFolders, paneFolders),
		focus:           paneFolders,
		selectionHolder: paneFolders,
		panes:           newPanes(keys, styles, location),
		status:          "",
	}
}

func (s Screen) Update(msg tea.Msg) outcome.Step {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s.pressed(msg)
	case look.Resized:
		return outcome.Stay(s.resized(msg.Box))
	case tea.BackgroundColorMsg:
		return outcome.Stay(s.withPanes(s.panes.withBackground(msg)))
	case SnippetsLoaded:
		return outcome.Stay(s.withPanes(s.panes.withSnippets(msg.Snippets, msg.Selecting)))
	case StatusShown:
		return outcome.Stay(s.withStatus(msg.Text))
	}

	return outcome.Stay(s)
}

func (s Screen) Received(received outcome.Outcome) outcome.Step {
	if _, ok := received.(outcome.SnippetRevealed); ok {
		return outcome.Stay(s.revealing()).Passing(received)
	}

	return outcome.Stay(s).Passing(received)
}

func (s Screen) View() string {
	return s.ViewUnder(s.ShortHelp())
}

func (s Screen) ViewUnder(hints []key.Binding) string {
	return s.panesView() + "\n" + statusLine(s.styles, s.status, s.hint(hints), s.box.Width)
}

func (s Screen) ShortHelp() []key.Binding {
	if s.layout.single {
		return nil
	}

	return s.panes.keyMaps()[s.focus].ShortHelp()
}

func (s Screen) FullHelp() [][]key.Binding {
	return s.panes.keyMaps()[s.focus].FullHelp()
}

func (s Screen) pressed(msg tea.KeyPressMsg) outcome.Step {
	switch {
	case s.global.Matches(msg, binding.Quit):
		return outcome.Stay(s).Passing(outcome.QuitAsked{})
	case s.global.Matches(msg, binding.NewSnippet):
		return s.opening(editoverlay.New(s.keys, s.styles))
	case s.global.Matches(msg, binding.Search):
		return s.opening(searchpopup.New(s.keys, s.styles, s.panes.preview.Cleared(), s.panes.list.Snippets()))
	}

	if navigate, ok := navigationFor(s.global, msg); ok {
		return outcome.Stay(s.focusedOn(navigate(s.focus, s.selectionHolder)))
	}

	return s.focusedPressed(msg)
}

func (s Screen) opening(child outcome.Overlay, cmd tea.Cmd) outcome.Step {
	return outcome.Stay(s).Opening(child).Running(cmd)
}

func (s Screen) focusedPressed(msg tea.KeyPressMsg) outcome.Step {
	pressed, outcomes, cmd := s.panes.pressed(s.focus, msg)

	return outcome.Stay(s.withPanes(pressed)).Passing(outcomes...).Running(cmd)
}

func (s Screen) withPanes(next panes) Screen {
	s.panes = next

	return s
}

func (s Screen) withStatus(text string) Screen {
	s.status = text

	return s
}

func (s Screen) focusedOn(target pane) Screen {
	next := s
	next.focus = target

	return next.arranged()
}

func (s Screen) revealing() Screen {
	next := s
	next.focus = paneSnippet
	next.selectionHolder = paneFolders

	return next.arranged()
}

func (s Screen) resized(box look.Size) Screen {
	next := s
	next.box = box

	return next.arranged()
}

func (s Screen) arranged() Screen {
	s.layout = arrange(s.box, s.focus, s.tallLeft())
	s.panes = s.panes.resized(s.layout)

	return s
}

func (s Screen) tallLeft() pane {
	if s.focus.inLeftColumn() {
		return s.focus
	}

	return s.selectionHolder
}

func (s Screen) hint(hints []key.Binding) string {
	if s.layout.single && len(hints) == 0 {
		return tooSmallHint
	}

	return hintFor(hints, s.box.Width/hintWidthDivisor)
}

func (s Screen) panesView() string {
	if s.layout.single {
		return s.paneFrame(s.focus)
	}

	left := lipgloss.JoinVertical(lipgloss.Left, s.paneFrame(paneFolders), s.paneFrame(paneTags))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, s.paneFrame(paneList), s.paneFrame(paneSnippet))
}

func (s Screen) paneFrame(p pane) string {
	frame := s.paneStyle(p)

	return look.Frame(frame, p.title(), s.panes.body(p, frame), s.layout.of(p))
}

func (s Screen) paneStyle(p pane) look.FrameStyle {
	if p == s.focus {
		return s.styles.Focused
	}

	return s.styles.Unfocused
}
