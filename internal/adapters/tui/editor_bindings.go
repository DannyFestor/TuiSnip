package tui

import "charm.land/bubbles/v2/key"

const (
	labelSave   = "save"
	labelCancel = "cancel"
	labelField  = "field"
	labelLeave  = "leave"
)

type editorBindings struct {
	save         key.Binding
	cancel       key.Binding
	nextField    key.Binding
	prevField    key.Binding
	openField    key.Binding
	contentSave  key.Binding
	contentLeave key.Binding
}

func newEditorBindings(editor EditorKeyMap, content ContentKeyMap) editorBindings {
	return editorBindings{
		save:         labelled(editor.Save, labelSave),
		cancel:       labelled(editor.Cancel, labelCancel),
		nextField:    labelled(editor.NextField, labelField),
		prevField:    unlabelled(editor.PrevField),
		openField:    unlabelled(editor.OpenField),
		contentSave:  labelled(content.Save, labelSave),
		contentLeave: labelled(content.Leave, labelLeave),
	}
}

func (b editorBindings) fieldHints() []key.Binding {
	return []key.Binding{b.save, b.cancel, b.nextField}
}

func (b editorBindings) contentHints() []key.Binding {
	return []key.Binding{b.contentSave, b.contentLeave}
}
