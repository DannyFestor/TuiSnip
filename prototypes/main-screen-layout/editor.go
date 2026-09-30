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
	inputs [fieldLanguage]textinput.Model
	lang   string
	body   textarea.Model
}

func newEditor(sn *snippet) *editor {
	e := &editor{target: sn, isNew: sn.id == "", lang: sn.lang, body: textarea.New()}
	values := [fieldLanguage]string{sn.title, sn.desc, strings.Join(sn.tags, ", ")}
	placeholders := [fieldLanguage]string{"required", "optional", "comma-separated"}
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
	switch {
	case f < fieldLanguage:
		e.inputs[f].Focus()
	case f == fieldContent:
		e.body.Focus()
	}
}

func (e *editor) tags() []string {
	var out []string
	for t := range strings.SplitSeq(e.inputs[fieldTags].Value(), ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (e *editor) dirty() bool {
	sn := e.target
	return e.isNew ||
		e.inputs[fieldTitle].Value() != sn.title ||
		e.inputs[fieldDescription].Value() != sn.desc ||
		!slices.Equal(e.tags(), sn.tags) ||
		e.lang != sn.lang ||
		e.body.Value() != sn.content
}

func (e *editor) update(m *model, msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := e.binding(m, key.String()); handled {
			return cmd
		}
	}
	var cmd tea.Cmd
	switch {
	case e.field < fieldLanguage:
		e.inputs[e.field], cmd = e.inputs[e.field].Update(msg)
	case e.field == fieldContent:
		e.body, cmd = e.body.Update(msg)
	}
	return cmd
}

func (e *editor) binding(m *model, key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+s":
		e.save(m)
	case "esc":
		m.confirmIfDirty("Discard unsaved changes?", func() tea.Cmd { m.editor = nil; return nil })
	case "ctrl+l":
		e.pickLanguage(m)
	case "ctrl+t":
		e.focusField(fieldTags)
	case "ctrl+e":
		m.status = "Would suspend and open the content in $EDITOR, then land it here as an unsaved change"
	case "tab":
		if e.field == fieldContent {
			e.body.InsertString("\t")
			return nil, true
		}
		e.focusField(e.field + 1)
	case "shift+tab":
		e.focusField((e.field + fieldCount - 1) % fieldCount)
	case "enter":
		if e.field == fieldContent {
			return nil, false
		}
		if e.field == fieldLanguage {
			e.pickLanguage(m)
			return nil, true
		}
		e.focusField(e.field + 1)
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
	sn.title, sn.desc, sn.tags, sn.lang, sn.content = title, e.inputs[fieldDescription].Value(), e.tags(), e.lang, e.body.Value()
	sn.updated = time.Now()
	if e.isNew {
		sn.id = m.store.newID()
		m.store.snippets = append(m.store.snippets, sn)
	}
	m.store.ensureTags(sn.tags)
	m.editor = nil
	m.status = "Saved"
}
