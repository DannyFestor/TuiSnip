// PROTOTYPE — throwaway. Edit mode of the Snippet pane, shared by every variant.
package main

import (
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const (
	fieldTitle = iota
	fieldDescription
	fieldTags
	fieldLanguage
	fieldContent
	fieldCount
)

var fieldNames = [...]string{"Title", "Description", "Tags", "Language", "Content"}

type editor struct {
	target *snippet
	isNew  bool
	field  int
	// inBody is false while Content is the selected field but its textarea isn't entered.
	inBody bool
	inputs [fieldTags]textinput.Model
	tags   []string
	lang   string
	body   textarea.Model
	// locked holds content with tabs, which the stock textarea would rewrite as spaces.
	locked bool
}

func newEditor(sn *snippet) *editor {
	e := &editor{target: sn, isNew: sn.id == "", lang: sn.lang, tags: slices.Clone(sn.tags), body: textarea.New()}
	e.locked = strings.Contains(sn.content, "\t")
	values := [fieldTags]string{sn.title, sn.desc}
	placeholders := [fieldTags]string{"required", "optional"}
	for i := range e.inputs {
		in := textinput.New()
		in.Prompt = ""
		in.Placeholder = placeholders[i]
		in.SetValue(values[i])
		e.inputs[i] = in
	}
	e.body.SetValue(sn.content)
	e.body.MoveToBegin()
	e.focusField(fieldTitle)
	return e
}

func (e *editor) focusField(f int) {
	e.field = f
	for i := range e.inputs {
		e.inputs[i].Blur()
	}
	e.body.Blur()
	e.inBody = false
	switch {
	case f < fieldTags:
		e.inputs[f].Focus()
	case f == fieldContent && !e.locked:
		e.inBody = true
		e.body.Focus()
	}
}

func (e *editor) content() string {
	if e.locked {
		return e.target.content
	}
	return e.body.Value()
}

func (e *editor) leaveBody() {
	e.inBody = false
	e.body.Blur()
}

// dedent removes one indent level (a tab, or up to four spaces) from the cursor's line.
func (e *editor) dedent() {
	lines := strings.Split(e.body.Value(), "\n")
	row, col := e.body.Line(), e.body.Column()
	line := lines[row]
	trimmed := strings.TrimPrefix(line, "\t")
	if trimmed == line {
		trimmed = strings.TrimPrefix(line, strings.Repeat(" ", min(4, len(line)-len(strings.TrimLeft(line, " ")))))
	}
	removed := len(line) - len(trimmed)
	if removed == 0 {
		return
	}
	lines[row] = trimmed
	e.body.SetValue(strings.Join(lines, "\n"))
	e.body.MoveToBegin()
	for range row {
		e.body.CursorDown()
	}
	e.body.SetCursorColumn(max(0, col-removed))
}

func (e *editor) dirty() bool {
	sn := e.target
	return e.isNew ||
		e.inputs[fieldTitle].Value() != sn.title ||
		e.inputs[fieldDescription].Value() != sn.desc ||
		!slices.Equal(e.tags, sn.tags) ||
		e.lang != sn.lang ||
		e.content() != sn.content
}

func (e *editor) update(m *model, msg tea.Msg) tea.Cmd {
	if paste, ok := msg.(tea.PasteMsg); ok && e.inBody && strings.Contains(paste.Content, "\t") {
		m.status = "Pasted text contains tabs; use ctrl+e to edit in $EDITOR"
		return nil
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := e.binding(m, key.String()); handled {
			return cmd
		}
	}
	var cmd tea.Cmd
	switch {
	case e.field < fieldTags:
		e.inputs[e.field], cmd = e.inputs[e.field].Update(msg)
	case e.inBody:
		e.body, cmd = e.body.Update(msg)
	}
	return cmd
}

func (e *editor) binding(m *model, key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+s":
		e.save(m)
	case "ctrl+l":
		e.pickLanguage(m)
	case "ctrl+t":
		e.focusField(fieldTags)
		m.overlay = newTagEditor(m, e)
	case "ctrl+e":
		m.status = "Would suspend and open the content in $EDITOR, then land it here as an unsaved change"
	default:
		if e.inBody {
			return e.bodyBinding(key)
		}
		return e.fieldBinding(m, key)
	}
	return nil, true
}

func (e *editor) bodyBinding(key string) (tea.Cmd, bool) {
	switch key {
	case "esc":
		e.leaveBody()
	case "tab":
		e.body.InsertString("\t")
	case "shift+tab":
		e.dedent()
	case "up":
		if e.body.Line() > 0 {
			return nil, false
		}
		e.focusField(fieldLanguage)
	default:
		return nil, false
	}
	return nil, true
}

func (e *editor) fieldBinding(m *model, key string) (tea.Cmd, bool) {
	switch key {
	case "esc":
		m.confirmIfDirty("Discard unsaved changes?", func() tea.Cmd { m.editor = nil; return nil })
	case "up", "shift+tab":
		e.focusField(max(0, e.field-1))
	case "down", "tab":
		e.focusField(min(fieldContent, e.field+1))
	case "enter":
		switch e.field {
		case fieldLanguage:
			e.pickLanguage(m)
		case fieldTags:
			m.overlay = newTagEditor(m, e)
		case fieldContent:
			e.focusField(fieldContent)
		default:
			e.focusField(min(fieldContent, e.field+1))
		}
	default:
		return nil, false
	}
	return nil, true
}

func (e *editor) pickLanguage(m *model) {
	m.overlay = newPicker("Language", languages, func(choice string) { e.lang = choice })
}

func (e *editor) save(m *model) {
	title := strings.TrimSpace(e.inputs[fieldTitle].Value())
	if title == "" {
		m.status = "Title is required"
		e.focusField(fieldTitle)
		return
	}
	sn := e.target
	sn.title, sn.desc, sn.tags, sn.lang, sn.content = title, e.inputs[fieldDescription].Value(), slices.Clone(e.tags), e.lang, e.content()
	sn.updated = time.Now()
	if e.isNew {
		sn.id = m.store.newID()
		m.store.snippets = append(m.store.snippets, sn)
	}
	m.store.ensureTags(sn.tags)
	m.editor = nil
	m.status = "Saved"
}
