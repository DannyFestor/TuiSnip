package snippetpane

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/fragmentstrip"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	dateLayout       = "2006-01-02"
	lineNumberFormat = "%4d │ "
	blankGutter      = "     │ "
	wrappedGutter    = "   ↪ │ "
	metaSeparator    = " · "
	tagMarker        = "#"
	tagSeparator     = " "
	fixedHeaderLines = 4
)

type Pane struct {
	global     binding.Set
	keys       binding.Set
	emptyHints []key.Binding
	styles     look.Styles
	snippet    domain.Snippet
	shown      bool
	trashed    Trashed
	strip      fragmentstrip.Strip
	viewport   viewport.Model
	location   *time.Location
	paths      folderpath.Paths
	box        look.Size
}

func New(keys binding.Keys, styles look.Styles, location *time.Location) Pane {
	code := viewport.New()
	code.LeftGutterFunc = lineNumbers(styles)

	return Pane{
		global:     keys.For(binding.ScopeGlobal),
		keys:       keys.For(binding.ScopeSnippetPane),
		emptyHints: keys.EmptyListHints(),
		styles:     styles,
		snippet:    domain.Snippet{},
		shown:      false,
		trashed:    Trashed{Place: "", DeletedAt: time.Time{}},
		strip:      fragmentstrip.New(styles),
		viewport:   code,
		location:   location,
		paths:      folderpath.Paths{},
		box:        look.Size{Width: 0, Height: 0},
	}
}

//nolint:unparam // every embedded child returns a Cmd (architecture.md), so parents handle each child alike.
func (p Pane) Update(msg tea.Msg) (Pane, []outcome.Outcome, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.pressed(msg), p.copyRequested(msg), nil
	case look.Resized:
		return p.resized(msg.Box), nil, nil
	case look.Restyled:
		return p.restyled(msg.Styles), nil, nil
	case pointer.Wheeled:
		return p.wheeled(msg), nil, nil
	case pointer.Clicked:
		return p.clicked(msg), nil, nil
	}

	return p, nil, nil
}

func (p Pane) View() string {
	if !p.shown {
		return look.EmptyHint(p.styles, p.emptyHints)
	}

	return strings.Join(append(p.header(), p.viewport.View()), "\n")
}

func (p Pane) ShortHelp() []key.Binding {
	return p.keys.ShortHelp()
}

func (p Pane) FullHelp() [][]key.Binding {
	return p.keys.FullHelp()
}

func (p Pane) Showing(snippet domain.Snippet) Pane {
	return p.displaying(snippet, Trashed{Place: "", DeletedAt: time.Time{}})
}

func (p Pane) ShowingInTrash(snippet domain.Snippet, trashed Trashed) Pane {
	return p.displaying(snippet, trashed)
}

func (p Pane) ShowingFragment(id domain.FragmentID) Pane {
	index, err := p.snippet.FragmentIndex(id)
	if err != nil {
		return p
	}

	return p.withStrip(p.strip.Showing(index))
}

func (p Pane) Shown() (domain.Snippet, bool) {
	return p.snippet, p.shown
}

func (p Pane) ShownFragment() (domain.Fragment, bool) {
	if !p.shown {
		return domain.Fragment{}, false
	}

	return p.fragmentInView(), true
}

func (p Pane) Copying() (outcome.CopyRequested, bool) {
	fragment, shown := p.ShownFragment()

	return outcome.CopyRequested{
		ID:         p.snippet.ID(),
		FragmentID: fragment.ID(),
		Title:      p.snippet.Title().String(),
		InTrash:    p.trashed.isSet(),
	}, shown
}

func (p Pane) WithPaths(paths folderpath.Paths) Pane {
	p.paths = paths

	return p
}

func (p Pane) Cleared() Pane {
	p.snippet = domain.Snippet{}
	p.shown = false
	p.viewport.SetContent("")

	return p
}

func (p Pane) displaying(snippet domain.Snippet, trashed Trashed) Pane {
	next := p
	next.snippet = snippet
	next.trashed = trashed
	next.strip = p.stripFor(snippet)

	if p.shown && p.snippet.ID() == snippet.ID() && p.snippet.UpdatedAt().Equal(snippet.UpdatedAt()) {
		return next
	}

	next.shown = true

	return next.rendered()
}

func (p Pane) pressed(msg tea.KeyPressMsg) Pane {
	switch {
	case p.keys.Matches(msg, binding.Wrap):
		return p.wrapToggled()
	case p.keys.Matches(msg, binding.NextFragment):
		return p.withStrip(p.strip.Next())
	case p.keys.Matches(msg, binding.PrevFragment):
		return p.withStrip(p.strip.Previous())
	}

	direction, ok := move.Pressed(p.global, msg)
	if !ok {
		return p
	}

	return p.scrolled(direction)
}

func (p Pane) copyRequested(msg tea.KeyPressMsg) []outcome.Outcome {
	request, ok := p.Copying()
	if !ok || !p.keys.Matches(msg, binding.Copy) {
		return nil
	}

	return []outcome.Outcome{request}
}

func (p Pane) clicked(click pointer.Clicked) Pane {
	if !p.shown || !p.strip.Visible() || click.At.Y != p.stripRow() {
		return p
	}

	return p.withStrip(p.strip.Clicked(click.At.X))
}

func (p Pane) stripFor(snippet domain.Snippet) fragmentstrip.Strip {
	if !p.shown || p.snippet.ID() != snippet.ID() {
		return p.strip.First().Of(snippet.Fragments())
	}

	return p.strip.Of(snippet.Fragments())
}

func (p Pane) withStrip(strip fragmentstrip.Strip) Pane {
	if !p.shown || strip.Shown() == p.strip.Shown() {
		return p
	}

	p.strip = strip

	return p.rendered()
}

func (p Pane) fragmentInView() domain.Fragment {
	return p.snippet.Fragments()[p.strip.Shown()]
}

func (p Pane) wrapToggled() Pane {
	p.viewport.SoftWrap = !p.viewport.SoftWrap

	return p
}

func (p Pane) restyled(styles look.Styles) Pane {
	next := p
	next.styles = styles
	next.strip = p.strip.Restyled(styles)
	next.viewport.LeftGutterFunc = lineNumbers(styles)

	if !next.shown {
		return next
	}

	return next.rendered()
}

func (p Pane) resized(box look.Size) Pane {
	next := p
	next.box = box

	return next.sized()
}

func (p Pane) scrolled(direction move.Direction) Pane {
	p.viewport = move.Scrolled(p.viewport, direction)

	return p
}

func (p Pane) wheeled(wheel pointer.Wheeled) Pane {
	p.viewport = move.Wheeled(p.viewport, wheel.Lines)

	return p
}

func (p Pane) rendered() Pane {
	fragment := p.fragmentInView()
	p.viewport.SetContent(look.Highlight(fragment.Content().String(), fragment.Language().String(), p.styles.CodeStyle))
	p.viewport.GotoTop()

	return p.sized()
}

func (p Pane) sized() Pane {
	p.viewport.SetWidth(p.box.Width)
	p.viewport.SetHeight(max(1, p.box.Height-p.headerHeight()))

	return p
}

func (p Pane) header() []string {
	lines := []string{
		p.styles.Bold.Render(p.snippet.Title().String()),
		p.styles.Dim.Render(p.meta()),
		p.styles.Dim.Render(p.timestamps()),
	}
	lines = append(lines, p.descriptionLines()...)
	lines = append(lines, p.styles.Dim.Render(strings.Repeat("─", p.box.Width)))

	return append(lines, p.stripLines()...)
}

func (p Pane) stripLines() []string {
	if !p.shown || !p.strip.Visible() {
		return nil
	}

	return []string{p.strip.View(p.box.Width)}
}

func (p Pane) stripRow() int {
	return fixedHeaderLines + len(p.descriptionLines())
}

func (p Pane) meta() string {
	place := p.place() + metaSeparator + p.fragmentInView().Language().String()

	tags := p.snippet.Tags()
	if len(tags) == 0 {
		return place
	}

	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tagMarker+tag.Name().String())
	}

	return place + metaSeparator + strings.Join(names, tagSeparator)
}

func (p Pane) headerHeight() int {
	return p.stripRow() + len(p.stripLines())
}

func (p Pane) descriptionLines() []string {
	description := p.snippet.Description().String()
	if !p.shown || description == "" {
		return nil
	}

	return strings.Split(description, "\n")
}

func (p Pane) place() string {
	if p.trashed.isSet() {
		return p.trashed.Place
	}

	return p.paths.Full(p.snippet.FolderID())
}

func (p Pane) timestamps() string {
	dates := "created " + p.date(p.snippet.CreatedAt()) + metaSeparator + "updated " + p.date(p.snippet.UpdatedAt())
	if !p.trashed.isSet() {
		return dates
	}

	return dates + metaSeparator + DeletedOn(p.trashed.DeletedAt, p.location)
}

func (p Pane) date(at time.Time) string {
	return at.In(p.location).Format(dateLayout)
}

func lineNumbers(styles look.Styles) viewport.GutterFunc {
	return func(line viewport.GutterContext) string {
		if line.Soft {
			return styles.Dim.Render(wrappedGutter)
		}

		if line.Index >= line.TotalLines {
			return styles.Dim.Render(blankGutter)
		}

		return styles.Dim.Render(fmt.Sprintf(lineNumberFormat, line.Index+1))
	}
}
