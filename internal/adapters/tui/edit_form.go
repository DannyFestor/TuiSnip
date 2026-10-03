package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
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
		title:       input.NewLine(""),
		description: input.NewLine(""),
		content:     input.NewContentArea(),
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

func (e editForm) resized(outer look.Size) editForm {
	inner := outer.Inner()
	inputWidth := max(1, inner.Width-len(fieldIndent)-fieldLabelWidth-cursorCell)

	e.title.SetWidth(inputWidth)
	e.description.SetWidth(inputWidth)
	e.content.SetWidth(inner.Width)
	e.content.SetHeight(max(1, inner.Height-fieldRows))

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
		return e.keys.content.ShortHelp()
	}

	return e.keys.fields.ShortHelp()
}

func (e editForm) externalEditorKey() string {
	return e.keys.content.FirstKey(binding.OpenInEditor)
}

func (e editForm) view(styles look.Styles, outer look.Size) string {
	lines := []string{
		e.fieldLine(styles, domain.FieldTitle, e.title.View()),
		e.fieldLine(styles, domain.FieldDescription, e.description.View()),
		e.fieldLine(styles, domain.FieldContent, e.contentEntryHint(styles)),
		e.content.View(),
	}

	return look.Frame(styles.Focused, e.frameTitle(), strings.Join(lines, "\n"), outer)
}

func (e editForm) fieldPressed(msg tea.KeyPressMsg) (editForm, editRequest, tea.Cmd) {
	fields := e.keys.fields

	switch {
	case fields.Matches(msg, binding.Save):
		return e, editSaves, nil
	case fields.Matches(msg, binding.Cancel):
		return e, editCancels, nil
	case fields.Matches(msg, binding.NextField), fields.Matches(msg, binding.OpenField):
		return e.advanced()
	case fields.Matches(msg, binding.PrevField):
		return e.steppedBack()
	}

	return e.typed(msg)
}

func (e editForm) contentPressed(msg tea.KeyPressMsg) (editForm, editRequest, tea.Cmd) {
	switch {
	case e.keys.content.Matches(msg, binding.Save):
		return e, editSaves, nil
	case e.keys.content.Matches(msg, binding.Leave):
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

func (e editForm) fieldLine(styles look.Styles, field domain.Field, entered string) string {
	cursor := fieldIndent
	label := styles.Plain

	if e.field == field {
		cursor = fieldCursor
		label = styles.Bold
	}

	if slices.Contains(e.invalid, field) {
		label = styles.Invalid
	}

	return cursor + label.Render(look.FitWidth(fieldLabel(field), fieldLabelWidth)) + entered
}

func (e editForm) contentEntryHint(styles look.Styles) string {
	if e.inContent {
		return ""
	}

	entryKeys := boundOnly(e.keys.fields.FirstKey(binding.OpenField), e.keys.fields.FirstKey(binding.NextField))

	if len(entryKeys) == 0 {
		return ""
	}

	return styles.Dim.Render(strings.Join(entryKeys, entryKeysJoiner) + entrySuffix)
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
