package editoverlay

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
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	overlayTitle    = "Editing"
	unsavedMarker   = " •"
	fieldCursor     = "› "
	fieldIndent     = "  "
	fieldLabelWidth = 12
	fieldRows       = 3
	tabCharacter    = "\t"
	lineBreak       = "\n"
	carriageReturn  = "\r"
	maxContentLines = 10_000
	entryKeysJoiner = " or "
	entrySuffix     = " to edit"
)

type form struct {
	keys        formKeys
	title       textinput.Model
	description textinput.Model
	content     textarea.Model
	readOnly    readOnlyContent
	original    entered
	field       domain.Field
	inContent   bool
	invalid     []domain.Field
}

func newForm(keys formKeys, original entered, readOnly readOnlyContent) (form, tea.Cmd) {
	filled := form{
		keys:        keys,
		title:       input.NewLine(""),
		description: input.NewLine(""),
		content:     input.NewContentArea(),
		readOnly:    readOnly,
		original:    original,
		field:       domain.FieldTitle,
		inContent:   false,
		invalid:     nil,
	}
	filled.title.SetValue(original.title)
	filled.description.SetValue(original.description)

	if !readOnly.held {
		filled.content.SetValue(original.content)
	}

	return filled.focused(domain.FieldTitle)
}

func (f form) update(msg tea.Msg) (form, request, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if f.inContent {
			return f.contentPressed(msg)
		}

		return f.fieldPressed(msg)
	case tea.PasteMsg:
		return f.pasted(msg)
	}

	return f, requestNothing, nil
}

func (f form) resized(outer look.Size) form {
	inner := outer.Inner()
	inputWidth := max(1, inner.Width-len(fieldIndent)-fieldLabelWidth-input.CursorWidth)

	f.title.SetWidth(inputWidth)
	f.description.SetWidth(inputWidth)
	f.content.SetWidth(inner.Width)
	f.content.SetHeight(max(1, inner.Height-fieldRows))

	return f
}

func (f form) restyled(codeStyle string) form {
	f.readOnly = f.readOnly.highlightedIn(codeStyle)

	return f
}

func (f form) withInvalid(fieldErrors []domain.FieldError) form {
	f.invalid = make([]domain.Field, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		f.invalid = append(f.invalid, fieldErr.Field)
	}

	return f
}

func (f form) changed() bool {
	return f.entered() != f.original
}

func (f form) entered() entered {
	content := f.content.Value()
	if f.readOnly.held {
		content = f.original.content
	}

	return entered{title: f.title.Value(), description: f.description.Value(), content: content}
}

func (f form) hints() []key.Binding {
	if f.inContent {
		return f.keys.content.ShortHelp()
	}

	return f.keys.fields.ShortHelp()
}

func (f form) externalEditorKey() string {
	return f.keys.content.FirstKey(binding.OpenInEditor)
}

func (f form) view(styles look.Styles, outer look.Size) string {
	lines := []string{
		f.fieldLine(styles, domain.FieldTitle, f.title.View()),
		f.fieldLine(styles, domain.FieldDescription, f.description.View()),
		f.fieldLine(styles, domain.FieldContent, f.contentEntryHint(styles)),
		f.contentView(),
	}

	return look.Frame(styles.Focused, f.frameTitle(), strings.Join(lines, "\n"), outer)
}

func (f form) contentView() string {
	if f.readOnly.held {
		return f.readOnly.highlighted
	}

	return f.content.View()
}

func (f form) fieldPressed(msg tea.KeyPressMsg) (form, request, tea.Cmd) {
	fields := f.keys.fields

	switch {
	case fields.Matches(msg, binding.Save):
		return f, requestSave, nil
	case fields.Matches(msg, binding.Cancel):
		return f, requestCancel, nil
	case fields.Matches(msg, binding.NextField), fields.Matches(msg, binding.OpenField):
		return f.advanced()
	case fields.Matches(msg, binding.PrevField):
		return f.steppedBack()
	}

	return f.typed(msg)
}

func (f form) contentPressed(msg tea.KeyPressMsg) (form, request, tea.Cmd) {
	switch {
	case f.keys.content.Matches(msg, binding.Save):
		return f, requestSave, nil
	case f.keys.content.Matches(msg, binding.Leave):
		return f.leftContent(), requestNothing, nil
	case f.keys.content.Matches(msg, binding.Indent):
		f.content = indentedLine(f.content)

		return f, requestNothing, nil
	case f.keys.content.Matches(msg, binding.Dedent):
		var cmd tea.Cmd

		f.content, cmd = dedentedLine(f.content)

		return f, requestNothing, cmd
	case f.leavesContentUpward(msg):
		return f.steppedBack()
	}

	return f.typed(msg)
}

func (f form) leavesContentUpward(msg tea.KeyPressMsg) bool {
	return f.content.Line() == 0 && key.Matches(msg, f.content.KeyMap.LinePrevious)
}

func (f form) pasted(msg tea.PasteMsg) (form, request, tea.Cmd) {
	if !f.inContent {
		return f.typed(msg)
	}

	switch {
	case strings.Contains(msg.Content, tabCharacter):
		return f, requestRefusePasteWithTabs, nil
	case f.linesAfterPaste(msg.Content) > maxContentLines:
		return f, requestRefuseOverlongPaste, nil
	}

	return f.typed(msg)
}

func (f form) linesAfterPaste(pasted string) int {
	return f.content.LineCount() - lineBreaksOnInsert(f.content.SelectedText()) + lineBreaksOnInsert(pasted)
}

// The textarea turns every carriage return into a line break before it inserts text, so "\r\n" becomes two.
func lineBreaksOnInsert(text string) int {
	return strings.Count(text, lineBreak) + strings.Count(text, carriageReturn)
}

func (f form) typed(msg tea.Msg) (form, request, tea.Cmd) {
	var cmd tea.Cmd

	switch f.field {
	case domain.FieldTitle:
		f.title, cmd = f.title.Update(msg)
	case domain.FieldDescription:
		f.description, cmd = f.description.Update(msg)
	case domain.FieldContent, domain.FieldLanguage:
		if f.inContent {
			f.content, cmd = f.content.Update(msg)
		}
	case domain.FieldFolderName, domain.FieldTagName:
	}

	return f, requestNothing, cmd
}

func (f form) advanced() (form, request, tea.Cmd) {
	if f.field == domain.FieldContent {
		if f.readOnly.held {
			return f, requestNothing, nil
		}

		next, cmd := f.enteredContent()

		return next, requestNothing, cmd
	}

	next, cmd := f.focused(f.fieldAt(1))

	return next, requestNothing, cmd
}

func (f form) steppedBack() (form, request, tea.Cmd) {
	next, cmd := f.focused(f.fieldAt(-1))

	return next, requestNothing, cmd
}

func (f form) focused(field domain.Field) (form, tea.Cmd) {
	f.field = field
	f = f.leftContent()
	f.title.Blur()
	f.description.Blur()

	switch field {
	case domain.FieldTitle:
		return f, f.title.Focus()
	case domain.FieldDescription:
		return f, f.description.Focus()
	case domain.FieldContent, domain.FieldLanguage, domain.FieldFolderName, domain.FieldTagName:
	}

	return f, nil
}

func (f form) enteredContent() (form, tea.Cmd) {
	f.inContent = true

	return f, f.content.Focus()
}

func (f form) leftContent() form {
	f.inContent = false
	f.content.Blur()

	return f
}

func (f form) fieldAt(offset int) domain.Field {
	fields := editFields()
	index := slices.Index(fields, f.field) + offset

	return fields[max(0, min(index, len(fields)-1))]
}

func (f form) fieldLine(styles look.Styles, field domain.Field, entered string) string {
	cursor := fieldIndent
	label := styles.Plain

	if f.field == field {
		cursor = fieldCursor
		label = styles.Bold
	}

	if slices.Contains(f.invalid, field) {
		label = styles.Invalid
	}

	return cursor + label.Render(look.FitWidth(fieldLabel(field), fieldLabelWidth)) + entered
}

func (f form) contentEntryHint(styles look.Styles) string {
	if f.readOnly.held {
		return styles.Dim.Render(readOnlyText(f.externalEditorKey()))
	}

	if f.inContent {
		return ""
	}

	entryKeys := binding.BoundOnly(f.keys.fields.FirstKey(binding.OpenField), f.keys.fields.FirstKey(binding.NextField))

	if len(entryKeys) == 0 {
		return ""
	}

	return styles.Dim.Render(strings.Join(entryKeys, entryKeysJoiner) + entrySuffix)
}

func (f form) frameTitle() string {
	if f.changed() {
		return overlayTitle + unsavedMarker
	}

	return overlayTitle
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
	case domain.FieldFolderName, domain.FieldTagName:
	}

	return field.String()
}
