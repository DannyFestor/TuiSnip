// PROTOTYPE — throwaway. Overlays drawn over the main screen, shared by every variant.
package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
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
		{"editor", "ctrl+s save", "esc cancel", "tab/shift+tab field", "ctrl+l Language", "ctrl+e $EDITOR"},
	}
	cols := make([]string, len(sections))
	for i, s := range sections {
		cols[i] = lipgloss.NewStyle().MarginRight(3).Render(bold.Render(s[0]) + "\n" + strings.Join(s[1:], "\n"))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cols[:3]...) + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, cols[3:]...)
	return boxStyle.Render(body + "\n\n" + dim.Render(fmt.Sprintf("? or esc closes · focus: %s", paneNames[m.focus])))
}
