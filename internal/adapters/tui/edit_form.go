package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	editOverlayTitle = "Editing"
	unsavedMarker    = " •"
	fieldCursor      = "› "
	fieldIndent      = "  "
	fieldLabelWidth  = 12
	cursorCell       = 1
	fieldRows        = 3
	tabCharacter     = "\t"
	entryKeysJoiner  = " or "
	entrySuffix      = " to edit"
)

type editForm struct {
	keys        editorBindings
	title       textinput.Model
	description textinput.Model
	content     textarea.Model
	field       domain.Field
	inContent   bool
	invalid     []domain.Field
}

func newEditForm(keys editorBindings) (editForm, tea.Cmd) {
	form := editForm{
		keys:        keys,
		title:       newLineInput(""),
		description: newLineInput(""),
		content:     newContentArea(),
		field:       domain.FieldTitle,
		inContent:   false,
		invalid:     nil,
	}

	return form.focused(domain.FieldTitle)
}

func (e editForm) update(msg tea.Msg) (editForm, editRequest, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if e.inContent {
			return e.contentPressed(msg)
		}

		return e.fieldPressed(msg)
	case tea.PasteMsg:
		return e.pasted(msg)
	}

	return e, editStays, nil
}

func (e editForm) resized(outer size) editForm {
	inner := innerSize(outer)
	inputWidth := max(1, inner.width-len(fieldIndent)-fieldLabelWidth-cursorCell)

	e.title.SetWidth(inputWidth)
	e.description.SetWidth(inputWidth)
	e.content.SetWidth(inner.width)
	e.content.SetHeight(max(1, inner.height-fieldRows))

	return e
}

func (e editForm) withInvalid(fieldErrors []domain.FieldError) editForm {
	e.invalid = make([]domain.Field, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		e.invalid = append(e.invalid, fieldErr.Field)
	}

	return e
}

func (e editForm) changed() bool {
	return e.title.Value() != "" || e.description.Value() != "" || e.content.Value() != ""
}

func (e editForm) input() snippet.CreateInput {
	return snippet.CreateInput{
		Title:       e.title.Value(),
		Description: e.description.Value(),
		Content:     e.content.Value(),
	}
}

func (e editForm) hints() []key.Binding {
	if e.inContent {
		return e.keys.contentHints()
	}

	return e.keys.fieldHints()
}

func (e editForm) view(styles styleSet, outer size) string {
	lines := []string{
		e.fieldLine(styles, domain.FieldTitle, e.title.View()),
		e.fieldLine(styles, domain.FieldDescription, e.description.View()),
		e.fieldLine(styles, domain.FieldContent, e.contentEntryHint(styles)),
		e.content.View(),
	}

	return frame(styles.focused, e.frameTitle(), strings.Join(lines, "\n"), outer)
}

func (e editForm) fieldPressed(msg tea.KeyPressMsg) (editForm, editRequest, tea.Cmd) {
	switch {
	case key.Matches(msg, e.keys.save):
		return e, editSaves, nil
	case key.Matches(msg, e.keys.cancel):
		return e, editCancels, nil
	case key.Matches(msg, e.keys.nextField), key.Matches(msg, e.keys.openField):
		return e.advanced()
	case key.Matches(msg, e.keys.prevField):
		return e.steppedBack()
	}

	return e.typed(msg)
}

func (e editForm) contentPressed(msg tea.KeyPressMsg) (editForm, editRequest, tea.Cmd) {
	switch {
	case key.Matches(msg, e.keys.contentSave):
		return e, editSaves, nil
	case key.Matches(msg, e.keys.contentLeave):
		return e.leftContent(), editStays, nil
	case e.content.Line() == 0 && key.Matches(msg, e.content.KeyMap.LinePrevious):
		return e.steppedBack()
	}

	return e.typed(msg)
}

func (e editForm) pasted(msg tea.PasteMsg) (editForm, editRequest, tea.Cmd) {
	if e.inContent && strings.Contains(msg.Content, tabCharacter) {
		return e, editRefusesPaste, nil
	}

	return e.typed(msg)
}

func (e editForm) typed(msg tea.Msg) (editForm, editRequest, tea.Cmd) {
	var cmd tea.Cmd

	switch e.field {
	case domain.FieldTitle:
		e.title, cmd = e.title.Update(msg)
	case domain.FieldDescription:
		e.description, cmd = e.description.Update(msg)
	case domain.FieldContent, domain.FieldLanguage:
		if e.inContent {
			e.content, cmd = e.content.Update(msg)
		}
	}

	return e, editStays, cmd
}

func (e editForm) advanced() (editForm, editRequest, tea.Cmd) {
	if e.field == domain.FieldContent {
		next, cmd := e.enteredContent()

		return next, editStays, cmd
	}

	next, cmd := e.focused(e.fieldAt(1))

	return next, editStays, cmd
}

func (e editForm) steppedBack() (editForm, editRequest, tea.Cmd) {
	next, cmd := e.focused(e.fieldAt(-1))

	return next, editStays, cmd
}

func (e editForm) focused(field domain.Field) (editForm, tea.Cmd) {
	e.field = field
	e = e.leftContent()
	e.title.Blur()
	e.description.Blur()

	switch field {
	case domain.FieldTitle:
		return e, e.title.Focus()
	case domain.FieldDescription:
		return e, e.description.Focus()
	case domain.FieldContent, domain.FieldLanguage:
	}

	return e, nil
}

func (e editForm) enteredContent() (editForm, tea.Cmd) {
	e.inContent = true

	return e, e.content.Focus()
}

func (e editForm) leftContent() editForm {
	e.inContent = false
	e.content.Blur()

	return e
}

func (e editForm) fieldAt(offset int) domain.Field {
	fields := editFields()
	index := slices.Index(fields, e.field) + offset

	return fields[max(0, min(index, len(fields)-1))]
}

func (e editForm) fieldLine(styles styleSet, field domain.Field, entry string) string {
	cursor := fieldIndent
	label := styles.plain

	if e.field == field {
		cursor = fieldCursor
		label = styles.bold
	}

	if slices.Contains(e.invalid, field) {
		label = styles.invalid
	}

	return cursor + label.Render(fitWidth(fieldLabel(field), fieldLabelWidth)) + entry
}

func (e editForm) contentEntryHint(styles styleSet) string {
	if e.inContent {
		return ""
	}

	entryKeys := make([]string, 0, len(e.keys.openField.Keys())+len(e.keys.nextField.Keys()))
	for _, binding := range []key.Binding{e.keys.openField, e.keys.nextField} {
		if binding.Enabled() && len(binding.Keys()) > 0 {
			entryKeys = append(entryKeys, binding.Keys()[0])
		}
	}

	if len(entryKeys) == 0 {
		return ""
	}

	return styles.dim.Render(strings.Join(entryKeys, entryKeysJoiner) + entrySuffix)
}

func (e editForm) frameTitle() string {
	if e.changed() {
		return editOverlayTitle + unsavedMarker
	}

	return editOverlayTitle
}

func editFields() []domain.Field {
	return []domain.Field{domain.FieldTitle, domain.FieldDescription, domain.FieldContent}
}

func fieldLabel(field domain.Field) string {
	switch field {
	case domain.FieldTitle:
		return "Title"
	case domain.FieldDescription:
		return "Description"
	case domain.FieldContent:
		return "Content"
	case domain.FieldLanguage:
		return "Language"
	}

	return field.String()
}
