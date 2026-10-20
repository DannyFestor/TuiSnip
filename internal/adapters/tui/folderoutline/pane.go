package folderoutline

import (
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

const (
	countSeparator = " · "
	headerLines    = 5
	subfolderNoun  = "subfolder"
	snippetNoun    = "Snippet"
)

type Pane struct {
	global   binding.Set
	hints    []key.Binding
	styles   look.Styles
	location *time.Location
	header   []string
	viewport viewport.Model
	box      look.Size
}

func New(keys binding.Keys, styles look.Styles, location *time.Location) Pane {
	return Pane{
		global:   keys.For(binding.ScopeGlobal),
		hints:    keys.TrashedFolderHints(),
		styles:   styles,
		location: location,
		header:   nil,
		viewport: viewport.New(),
		box:      look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg), nil, nil
	case look.Resized:
		return p.resized(msg.Box), nil, nil
	case look.Restyled:
		p.styles = msg.Styles

		return p, nil, nil
	case pointer.Wheeled:
		p.viewport = move.Wheeled(p.viewport, msg.Lines)

		return p, nil, nil
	}

	return p, nil, nil
}

func (p Pane) View() string {
	lines := make([]string, 0, headerLines+1)
	for index, line := range p.header {
		lines = append(lines, p.headerStyle(index).Render(line))
	}

	lines = append(lines, p.styles.Dim.Render(strings.Repeat("─", p.box.Width)), p.viewport.View())

	return strings.Join(lines, "\n")
}

func (p Pane) ShortHelp() []key.Binding {
	return slices.Clone(p.hints)
}

func (Pane) FullHelp() [][]key.Binding {
	return nil
}

func (p Pane) Showing(trashed trash.Item, place string) Pane {
	next := p
	next.header = []string{
		trashed.Name(), place, countsOf(trashed.Contents), look.DeletedOn(trashed.DeletedAt, p.location),
	}
	next.viewport.SetContent(strings.Join(outlineLines(trashed.Contents.Outline, ""), "\n"))
	next.viewport.GotoTop()

	return next
}

func (p Pane) headerStyle(index int) lipgloss.Style {
	if index == 0 {
		return p.styles.Bold
	}

	return p.styles.Dim
}

func (p Pane) pressed(msg tea.KeyPressMsg) Pane {
	direction, ok := move.Pressed(p.global, msg)
	if !ok {
		return p
	}

	p.viewport = move.Scrolled(p.viewport, direction)

	return p
}

func (p Pane) resized(box look.Size) Pane {
	next := p
	next.box = box
	next.viewport.SetWidth(box.Width)
	next.viewport.SetHeight(max(1, box.Height-headerLines))

	return next
}

func countsOf(contents trash.Contents) string {
	return look.Counted(contents.SubfolderCount, subfolderNoun) + countSeparator +
		look.Counted(contents.SnippetCount, snippetNoun)
}
