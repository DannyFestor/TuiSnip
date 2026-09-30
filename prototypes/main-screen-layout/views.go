// PROTOTYPE — throwaway. Rendering shared by every variant; variants only arrange it.
package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/x/ansi"
)

var (
	accent   = lipgloss.Color("#7D56F4")
	muted    = lipgloss.Color("#666666")
	bold     = lipgloss.NewStyle().Bold(true)
	dim      = lipgloss.NewStyle().Foreground(muted)
	selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(accent)
	cursorBg = lipgloss.NewStyle().Background(lipgloss.Color("#3A3A3A"))
	boxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)
	pillBar  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFD75F"))
)

// box draws a pane of exactly w×h cells including its border.
func box(title string, body string, w, h int, focused bool) string {
	border := muted
	if focused {
		border = accent
	}
	title = fit(title, max(1, w-4))
	head := dim.Render(" " + title + " ")
	if focused {
		head = bold.Foreground(accent).Render(" " + title + " ")
	}
	inner := lipgloss.NewStyle().Height(max(0, h-3)).MaxHeight(max(0, h-3)).Render(clip(body, max(0, w-2)))
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).
		Width(w).Height(h).MaxHeight(h).Render(head + "\n" + inner)
}

// clip truncates every line instead of wrapping, so panes never grow taller than drawn.
func clip(s string, w int) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		l = ansi.Truncate(l, w, "…")
		lines[i] = l + strings.Repeat(" ", max(0, w-ansi.StringWidth(l)))
	}
	return strings.Join(lines, "\n")
}

func fit(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func row(text, count string, w int, isCursor, focused bool) string {
	gap := max(1, w-lipgloss.Width(text)-lipgloss.Width(count))
	line := fit(text, w-lipgloss.Width(count)-1) + strings.Repeat(" ", gap) + dim.Render(count)
	line = lipgloss.NewStyle().Width(w).MaxWidth(w).Render(line)
	switch {
	case isCursor && focused:
		return selected.Width(w).Render(fit(text, w-lipgloss.Width(count)-1) + strings.Repeat(" ", gap) + count)
	case isCursor:
		return cursorBg.Render(line)
	}
	return line
}

func (m *model) entryRow(e entry, w int, isCursor bool) string {
	if m.inlineEdit != nil && isCursor {
		if m.inlineEdit.isNew {
			return row("  + "+m.inlineEdit.input.View(), "", w, false, false)
		}
		return row(strings.Repeat("  ", e.depth)+"✎ "+m.inlineEdit.input.View(), "", w, false, false)
	}
	var label string
	switch e.kind {
	case entryRoot:
		label = "◆ Root"
	case entryFolder:
		marker := "▾ "
		f := m.store.folder(e.id)
		if len(m.store.children(e.id)) == 0 {
			marker = "  "
		} else if f.collapsed {
			marker = "▸ "
		}
		label = strings.Repeat("  ", e.depth-1) + marker + f.name
	default:
		label = "# " + e.id
	}
	return row(label, fmt.Sprint(m.entryCount(e)), w, isCursor, m.focus == paneSidebar)
}

// sidebarRows renders the Folder tree and Tag list; section filters to one of them (-1 = both).
func (m *model) sidebarRows(w int, section entryKind, cursorOn bool) string {
	var b strings.Builder
	sel, _ := m.selectedEntry()
	wroteTagHeader := false
	for _, e := range m.entries() {
		isTag := e.kind == entryTag
		if section >= 0 && isTag != (section == entryTag) {
			continue
		}
		if isTag && !wroteTagHeader && section < 0 {
			b.WriteString("\n" + dim.Render("TAGS") + "\n")
			wroteTagHeader = true
		}
		b.WriteString(m.entryRow(e, w, cursorOn && e == sel) + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m *model) listRows(w, h int) string {
	var b strings.Builder
	if m.searching && !m.variant().searchAsPalette() {
		b.WriteString(m.searchLine(w) + "\n")
		h--
	}
	list := m.listed()
	if len(list) == 0 {
		return b.String() + m.emptyHint()
	}
	start := max(0, m.listCursor-h+1)
	for i := start; i < len(list) && i < start+h; i++ {
		sn := list[i]
		meta := m.snippetMeta(sn)
		b.WriteString(row(sn.title, meta, w, i == m.listCursor, m.focus == paneList) + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m *model) snippetMeta(sn *snippet) string {
	switch {
	case m.searching:
		return strings.TrimPrefix(m.store.path(sn.folder), "Root / ")
	case m.sort == sortUpdated:
		return sn.updated.Format("Jan 2")
	case m.sort == sortCreated:
		return sn.created.Format("Jan 2")
	}
	return sn.lang
}

func (m *model) searchLine(w int) string {
	line := m.search.View()
	if !m.searchFocused {
		line = dim.Render("/ " + m.search.Value() + "  (/ to refine, esc in search clears)")
	}
	return lipgloss.NewStyle().Width(w).MaxWidth(w).Render(line)
}

func (m *model) emptyHint() string {
	if m.searching {
		return dim.Render("No Snippet matches.")
	}
	return dim.Render("No Snippets here.") + "\n\n" +
		"n  new Snippet\np  capture clipboard\n/  search\nN  new Folder\n?  help"
}

func (m *model) listTitle() string {
	if m.searching {
		return fmt.Sprintf("Search · %d results", len(m.listed()))
	}
	e, _ := m.selectedEntry()
	if e.kind == entryTag {
		return "# " + e.id + " · by " + sortNames[m.sort]
	}
	return m.store.path(m.browseFolder()) + " · by " + sortNames[m.sort]
}

func (m *model) snippetPaneTitle() string {
	if m.editor != nil {
		mark := ""
		if m.editor.dirty() {
			mark = " •"
		}
		if m.editor.isNew {
			return "New Snippet" + mark
		}
		return "Editing" + mark
	}
	return "Snippet"
}

func (m *model) snippetPaneBody(w, h int) string {
	if m.editor != nil {
		return m.editorBody(w, h)
	}
	sn := m.selectedSnippet()
	if sn == nil {
		if len(m.store.snippets) == 0 {
			return m.emptyHint()
		}
		return dim.Render("No Snippet selected.")
	}
	header := bold.Render(fit(sn.title, w)) + "\n" +
		dim.Render(fit(fmt.Sprintf("%s · %s · #%s", m.store.path(sn.folder), sn.lang, strings.Join(sn.tags, " #")), w)) + "\n" +
		dim.Render(fmt.Sprintf("created %s · updated %s", sn.created.Format("2006-01-02"), sn.updated.Format("2006-01-02 15:04")))
	if sn.desc != "" {
		header += "\n" + fit(sn.desc, w)
	}
	header += "\n" + dim.Render(strings.Repeat("─", w))
	codeHeight := max(1, h-lipgloss.Height(header))
	m.configureScroll(sn, w, codeHeight)
	return header + "\n" + m.scroll.View()
}

func (m *model) configureScroll(sn *snippet, w, h int) {
	if m.lastSelected != sn.id+sn.content {
		m.scroll.SetContent(highlight(sn))
		m.scroll.GotoTop()
		m.lastSelected = sn.id + sn.content
	}
	m.scroll.SetWidth(w)
	m.scroll.SetHeight(h)
	m.scroll.SoftWrap = m.wrap
	m.scroll.LeftGutterFunc = func(c viewport.GutterContext) string {
		switch {
		case c.Soft:
			return dim.Render("     │ ")
		case c.Index >= c.TotalLines:
			return dim.Render("   ~ │ ")
		}
		return dim.Render(fmt.Sprintf("%4d │ ", c.Index+1))
	}
}

func highlight(sn *snippet) string {
	var b strings.Builder
	// The viewport measures a tab as one cell; the terminal draws more, which overflows the pane.
	content := strings.ReplaceAll(strings.TrimSuffix(sn.content, "\n"), "\t", "    ")
	if err := quick.Highlight(&b, content, sn.lang, "terminal16m", "monokai"); err != nil {
		return sn.content
	}
	return b.String()
}

func (m *model) editorBody(w, h int) string {
	e := m.editor
	var b strings.Builder
	label := func(f int) string {
		name := fmt.Sprintf("%-12s", fieldNames[f])
		if f == e.field {
			return bold.Foreground(accent).Render("› " + name)
		}
		return dim.Render("  " + name)
	}
	for f := range fieldLanguage {
		e.inputs[f].SetWidth(max(10, w-16))
		b.WriteString(label(f) + e.inputs[f].View() + "\n")
	}
	b.WriteString(label(fieldLanguage) + e.lang + dim.Render("   (enter or ctrl+l to pick)") + "\n")
	contentHint := "enter or ↓ to edit"
	if e.inBody {
		contentHint = "esc leaves · tab indents · shift+tab dedents"
	}
	b.WriteString(label(fieldContent) + dim.Render(contentHint) + "\n")
	e.body.SetWidth(w)
	e.body.SetHeight(max(3, h-lipgloss.Height(b.String())-1))
	b.WriteString(e.body.View() + "\n")
	b.WriteString(dim.Render(fit("ctrl+s save · esc cancel · ↑/↓ field · ctrl+l Language · ctrl+t Tags · ctrl+e $EDITOR", w)))
	return b.String()
}

func (m *model) render() string {
	if m.width == 0 {
		return ""
	}
	bodyHeight := m.height - 2
	base := m.variant().layout(m, m.width, bodyHeight)
	base = lipgloss.NewStyle().Width(m.width).Height(bodyHeight).MaxHeight(bodyHeight).Render(base)
	screen := base + "\n" + m.statusLine() + "\n" + m.switcher()
	if m.overlay == nil {
		return screen
	}
	top := m.overlay.view(m)
	x := max(0, (m.width-lipgloss.Width(top))/2)
	y := max(0, (bodyHeight-lipgloss.Height(top))/2)
	return lipgloss.NewCompositor(lipgloss.NewLayer(screen), lipgloss.NewLayer(top).X(x).Y(y).Z(1)).Render()
}

func (m *model) statusLine() string {
	return clip(" "+m.status, m.width)
}

// switcher is the prototype's own chrome, deliberately loud so it is not judged as design.
func (m *model) switcher() string {
	mode := "browse"
	switch {
	case m.editor != nil:
		mode = "edit:" + fieldNames[m.editor.field]
	case m.searching:
		mode = "search"
	}
	sel, _ := m.selectedEntry()
	state := fmt.Sprintf("scope=%s  mode=%s  browse=%s  max=%v  wrap=%v", m.scope(), mode, m.entryName(sel), m.maximized, m.wrap)
	label := fmt.Sprintf(" PROTOTYPE  ~ ◀  %c — %s  ▶ `  │ %s ", 'A'+m.current, m.variant().name(), state)
	return pillBar.Render(clip(label, m.width))
}

func (m *model) scope() string {
	switch {
	case m.overlay != nil:
		if _, ok := m.overlay.(*confirm); ok {
			return "confirm"
		}
		return "picker"
	case m.inlineEdit != nil:
		return "sidebar(rename)"
	case m.editor != nil && m.focus == paneSnippet:
		return "editor"
	case m.searching && m.searchFocused:
		return "search"
	}
	return paneNames[m.focus]
}
