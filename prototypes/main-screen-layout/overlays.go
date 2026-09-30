// PROTOTYPE — throwaway. Overlays drawn over the main screen, shared by every variant.
package main

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type overlay interface {
	update(m *model, msg tea.KeyPressMsg) tea.Cmd
	view(m *model) string
}

type confirm struct {
	question string
	onYes    func() tea.Cmd
}

func newConfirm(question string, onYes func() tea.Cmd) *confirm {
	return &confirm{question: question, onYes: onYes}
}

func (c *confirm) update(m *model, msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "y":
		m.overlay = nil
		return c.onYes()
	case "n", "esc", "enter":
		m.overlay = nil
	}
	return nil
}

func (c *confirm) view(*model) string {
	return boxStyle.Width(min(70, len(c.question)+4)).Render(c.question + "\n\n" + dim.Render("y yes · n/esc/enter no"))
}

// picker is the shared filterable list behind the Language and Folder pickers.
type picker struct {
	title    string
	labels   []string
	values   []string
	disabled map[string]bool
	cursor   int
	filter   textinput.Model
	onPick   func(value string)
}

func newPicker(title string, options []string, onPick func(string)) *picker {
	return newLabelledPicker(title, options, options, nil, onPick)
}

func newLabelledPicker(title string, labels, values []string, disabled map[string]bool, onPick func(string)) *picker {
	f := textinput.New()
	f.Prompt = "filter: "
	f.Focus()
	return &picker{title: title, labels: labels, values: values, disabled: disabled, filter: f, onPick: onPick}
}

func newFolderPicker(m *model, title, movingFolder string, onPick func(string)) *picker {
	labels, values := []string{"Root"}, []string{""}
	disabled := map[string]bool{}
	var walk func(parent string, depth int)
	walk = func(parent string, depth int) {
		for _, f := range m.store.children(parent) {
			labels = append(labels, strings.Repeat("  ", depth)+f.name)
			values = append(values, f.id)
			if movingFolder != "" && m.store.isDescendant(f.id, movingFolder) {
				disabled[f.id] = true
			}
			walk(f.id, depth+1)
		}
	}
	walk("", 1)
	return newLabelledPicker(title, labels, values, disabled, onPick)
}

func (p *picker) visible() []int {
	var out []int
	for i, l := range p.labels {
		if matches(l, strings.TrimSpace(p.filter.Value()), true) {
			out = append(out, i)
		}
	}
	return out
}

func (p *picker) update(m *model, msg tea.KeyPressMsg) tea.Cmd {
	vis := p.visible()
	switch msg.String() {
	case "esc":
		m.overlay = nil
		return nil
	case "enter":
		if len(vis) > 0 && !p.disabled[p.values[vis[p.cursor]]] {
			m.overlay = nil
			p.onPick(p.values[vis[p.cursor]])
		}
		return nil
	case "up", "ctrl+p", "ctrl+k":
		p.cursor = max(0, p.cursor-1)
		return nil
	case "down", "ctrl+n", "ctrl+j":
		p.cursor = clamp(p.cursor+1, 0, max(0, len(vis)-1))
		return nil
	}
	var cmd tea.Cmd
	p.filter, cmd = p.filter.Update(msg)
	p.cursor = 0
	return cmd
}

func (p *picker) view(*model) string {
	var b strings.Builder
	b.WriteString(bold.Render(p.title) + "\n" + p.filter.View() + "\n\n")
	for n, i := range p.visible() {
		line := p.labels[i]
		switch {
		case p.disabled[p.values[i]]:
			line = dim.Render(line + "  (inside itself)")
		case n == p.cursor:
			line = selected.Render("› " + line)
		default:
			line = "  " + line
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + dim.Render("↑/↓ ctrl+n/p move · type to filter · enter pick · esc cancel"))
	return boxStyle.Width(46).Render(b.String())
}

type helpOverlay struct{}

func (helpOverlay) update(m *model, msg tea.KeyPressMsg) tea.Cmd {
	if k := msg.String(); k == "?" || k == "esc" {
		m.overlay = nil
	}
	return nil
}

func (helpOverlay) view(m *model) string {
	sections := [][]string{
		append([]string{"Navigation (" + m.variant().name() + ")"}, m.variant().navHelp()...),
		{"global", "q quit", "? help", "/ search", "z maximize", "n new Snippet", "p capture", "` ~ switch variant (prototype)"},
		{"sidebar", "N new Folder", "r rename", "d delete", "m move Folder", "space collapse"},
		{"snippet_list", "y copy", "e edit", "E $EDITOR", "m move", "c duplicate", "d delete", "s cycle sort"},
		{"snippet_pane", "y copy", "e edit", "E $EDITOR", "w wrap", "j/k pgup/pgdn scroll"},
		{"editor", "ctrl+s save", "esc leave Content, then cancel", "↑/↓ field (tab too)", "tab / shift+tab indent in Content", "ctrl+l Language", "ctrl+e $EDITOR"},
	}
	cols := make([]string, len(sections))
	for i, s := range sections {
		cols[i] = lipgloss.NewStyle().MarginRight(3).Render(bold.Render(s[0]) + "\n" + strings.Join(s[1:], "\n"))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cols[:3]...) + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, cols[3:]...)
	return boxStyle.Render(body + "\n\n" + dim.Render(fmt.Sprintf("? or esc closes · focus: %s", paneNames[m.focus])))
}

// searchPopup is Search as a centred popup in the Language picker's style.
type searchPopup struct {
	input   textinput.Model
	cursor  int
	preview viewport.Model
	shown   string
}


func newSearchPopup() *searchPopup {
	in := textinput.New()
	in.Prompt = "/ "
	in.Placeholder = "title, Tags, Description, content"
	in.Focus()
	return &searchPopup{input: in, preview: viewport.New()}
}

func (p *searchPopup) results(m *model) []*snippet {
	if strings.TrimSpace(p.input.Value()) == "" {
		return nil
	}
	return m.store.search(p.input.Value())
}

func (p *searchPopup) update(m *model, msg tea.KeyPressMsg) tea.Cmd {
	hits := p.results(m)
	switch msg.String() {
	case "esc":
		m.overlay = nil
		return nil
	case "enter":
		if len(hits) > 0 {
			m.overlay = nil
			m.revealSnippet(hits[p.cursor])
		}
		return nil
	case "up", "ctrl+p", "ctrl+k":
		p.cursor = max(0, p.cursor-1)
		return nil
	case "down", "ctrl+n", "ctrl+j":
		p.cursor = clamp(p.cursor+1, 0, max(0, len(hits)-1))
		return nil
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	p.cursor = 0
	return cmd
}

func (p *searchPopup) view(m *model) string {
	w, h := m.width*4/5, (m.height-2)*4/5
	resultsW := w * 2 / 5
	hits := p.results(m)
	var b strings.Builder
	b.WriteString(p.input.View() + "\n\n")
	if len(hits) == 0 {
		b.WriteString(dim.Render("Type to search every Snippet.") + "\n")
	}
	rows := h - 7
	start := max(0, p.cursor-rows+1)
	for i := start; i < len(hits) && i < start+rows; i++ {
		b.WriteString(row(hits[i].title, strings.TrimPrefix(m.store.path(hits[i].folder), "Root / "), resultsW-2, i == p.cursor, true) + "\n")
	}
	b.WriteString("\n" + dim.Render("↑/↓ move · enter reveal · esc close"))
	preview := dim.Render("No result selected.")
	if len(hits) > 0 {
		preview = m.snippetView(hits[p.cursor], &p.preview, &p.shown, w-resultsW-2, h-3)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		box(fmt.Sprintf("Search · %d results", len(hits)), b.String(), resultsW, h, true),
		box("Preview", preview, w-resultsW, h, false),
	)
}

// tagEditor toggles the edited Snippet's Tags; changes stay unsaved in the editor until ctrl+s.
type tagEditor struct {
	editor *editor
	filter textinput.Model
	cursor int
}

type tagRow struct {
	name   string
	count  int
	create bool
}

func newTagEditor(_ *model, e *editor) *tagEditor {
	f := textinput.New()
	f.Prompt = "filter or new Tag: "
	f.Focus()
	return &tagEditor{editor: e, filter: f}
}

func (t *tagEditor) rows(m *model) []tagRow {
	query := strings.TrimSpace(t.filter.Value())
	names := slices.Clone(m.store.tags)
	for _, tag := range t.editor.tags {
		if !containsFold(names, tag) {
			names = append(names, tag)
		}
	}
	var out []tagRow
	for _, n := range names {
		if matches(n, query, true) {
			out = append(out, tagRow{name: n, count: m.store.tagCount(n)})
		}
	}
	if query != "" && !containsFold(names, query) {
		out = append(out, tagRow{name: query, create: true})
	}
	return out
}

func containsFold(list []string, s string) bool {
	return slices.ContainsFunc(list, func(x string) bool { return strings.EqualFold(x, s) })
}

func (t *tagEditor) toggle(name string) {
	e := t.editor
	if i := slices.IndexFunc(e.tags, func(x string) bool { return strings.EqualFold(x, name) }); i >= 0 {
		e.tags = slices.Delete(e.tags, i, i+1)
		return
	}
	e.tags = append(e.tags, name)
}

func (t *tagEditor) update(m *model, msg tea.KeyPressMsg) tea.Cmd {
	rows := t.rows(m)
	switch msg.String() {
	case "esc":
		m.overlay = nil
		return nil
	case "enter":
		t.accept(m, rows)
		return nil
	case "up", "ctrl+p", "ctrl+k":
		t.cursor = max(0, t.cursor-1)
		return nil
	case "down", "ctrl+n", "ctrl+j":
		t.cursor = clamp(t.cursor+1, 0, max(0, len(rows)-1))
		return nil
	}
	var cmd tea.Cmd
	t.filter, cmd = t.filter.Update(msg)
	t.cursor = 0
	return cmd
}

func (t *tagEditor) accept(m *model, rows []tagRow) {
	if len(rows) == 0 {
		return
	}
	r := rows[t.cursor]
	if r.create && strings.Contains(r.name, ",") {
		m.status = "Tag names cannot contain commas"
		return
	}
	t.toggle(r.name)
	if r.create {
		t.filter.Reset()
		t.cursor = 0
	}
}

func (t *tagEditor) view(m *model) string {
	const w = 44
	var b strings.Builder
	b.WriteString(bold.Render("Tags") + dim.Render("  "+strings.Join(t.editor.tags, ", ")) + "\n" + t.filter.View() + "\n\n")
	for i, r := range t.rows(m) {
		mark := "  "
		if containsFold(t.editor.tags, r.name) {
			mark = "✓ "
		}
		label, count := mark+r.name, fmt.Sprint(r.count)
		if r.create {
			label, count = fmt.Sprintf("+ create %q", r.name), "new"
		}
		b.WriteString(row(label, count, w, i == t.cursor, true) + "\n")
	}
	b.WriteString("\n" + dim.Render("enter toggle / create · ↑/↓ move · esc done"))
	return boxStyle.Width(w + 6).Render(b.String())
}
