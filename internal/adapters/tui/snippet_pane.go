package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	snippetPaneTitle = "4 Snippet"
	dateLayout       = "2006-01-02"
	lineNumberFormat = "%4d │ "
	blankGutter      = "     │ "
	metaSeparator    = " · "
	fixedHeaderLines = 4
)

type snippetPane struct {
	snippet   domain.Snippet
	shown     bool
	viewport  viewport.Model
	location  *time.Location
	codeStyle string
	inner     size
}

func newSnippetPane(styles styleSet, location *time.Location) snippetPane {
	code := viewport.New()
	code.LeftGutterFunc = lineNumbers(styles)

	return snippetPane{
		snippet:   domain.Snippet{},
		shown:     false,
		viewport:  code,
		location:  location,
		codeStyle: darkCodeStyle,
		inner:     size{width: 0, height: 0},
	}
}

func (p snippetPane) showing(snippet domain.Snippet) snippetPane {
	if p.shown && p.snippet.ID() == snippet.ID() && p.snippet.UpdatedAt().Equal(snippet.UpdatedAt()) {
		return p
	}

	next := p
	next.snippet = snippet
	next.shown = true

	return next.rendered()
}

func (p snippetPane) cleared() snippetPane {
	p.snippet = domain.Snippet{}
	p.shown = false
	p.viewport.SetContent("")

	return p
}

func (p snippetPane) withCodeStyle(codeStyle string) snippetPane {
	next := p
	next.codeStyle = codeStyle

	if !next.shown {
		return next
	}

	return next.rendered()
}

func (p snippetPane) resized(inner size) snippetPane {
	next := p
	next.inner = inner

	return next.sized()
}

func (p snippetPane) moved(move movement) snippetPane {
	switch move {
	case moveDown:
		p.viewport.ScrollDown(1)
	case moveUp:
		p.viewport.ScrollUp(1)
	case moveTop:
		p.viewport.GotoTop()
	case moveBottom:
		p.viewport.GotoBottom()
	case movePageDown:
		p.viewport.PageDown()
	case movePageUp:
		p.viewport.PageUp()
	case moveNone:
	}

	return p
}

func (p snippetPane) body(styles styleSet, hints []key.Binding) string {
	if !p.shown {
		return emptyHint(styles, hints)
	}

	return strings.Join(append(p.header(styles), p.viewport.View()), "\n")
}

func (p snippetPane) rendered() snippetPane {
	fragment := p.snippet.FirstFragment()
	p.viewport.SetContent(highlight(fragment.Content().String(), fragment.Language().String(), p.codeStyle))
	p.viewport.GotoTop()

	return p.sized()
}

func (p snippetPane) sized() snippetPane {
	p.viewport.SetWidth(p.inner.width)
	p.viewport.SetHeight(max(1, p.inner.height-p.headerHeight()))

	return p
}

func (p snippetPane) header(styles styleSet) []string {
	lines := []string{
		styles.bold.Render(p.snippet.Title().String()),
		styles.dim.Render(rootPath + metaSeparator + p.snippet.FirstFragment().Language().String()),
		styles.dim.Render(p.timestamps()),
	}
	lines = append(lines, p.descriptionLines()...)

	return append(lines, styles.dim.Render(strings.Repeat("─", p.inner.width)))
}

func (p snippetPane) headerHeight() int {
	return fixedHeaderLines + len(p.descriptionLines())
}

func (p snippetPane) descriptionLines() []string {
	description := p.snippet.Description().String()
	if !p.shown || description == "" {
		return nil
	}

	return strings.Split(description, "\n")
}

func (p snippetPane) timestamps() string {
	return "created " + p.snippet.CreatedAt().In(p.location).Format(dateLayout) +
		metaSeparator + "updated " + p.snippet.UpdatedAt().In(p.location).Format(dateLayout)
}

func lineNumbers(styles styleSet) viewport.GutterFunc {
	return func(line viewport.GutterContext) string {
		if line.Soft || line.Index >= line.TotalLines {
			return styles.dim.Render(blankGutter)
		}

		return styles.dim.Render(fmt.Sprintf(lineNumberFormat, line.Index+1))
	}
}
