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
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	overlayTitle      = "Editing"
	unsavedMarker     = " •"
	fieldCursor       = "› "
	fieldIndent       = "  "
	fieldLabelWidth   = 12
	fieldRows         = 5
	tagPrefix         = "#"
	tagSeparator      = " "
	tabCharacter      = "\t"
	lineBreak         = "\n"
	carriageReturn    = "\r"
	maxContentLines   = 10_000
	emptyContentLines = 1
	entryKeysJoiner   = " or "
	entrySuffix       = " to edit"
	pickSuffix        = " to pick"
	keysHintOpen      = "   ("
	keysHintClose     = ")"
)

type form struct {
	keys        formKeys
	title       textinput.Model
	description textinput.Model
	tags        tagchoice.Chosen
	language    value.Language
	content     textarea.Model
	readOnly    readOnlyContent
	original    entered
	field       domain.Field
	inContent   bool
	invalid     []domain.Field
}

func newForm(keys formKeys, styles look.Styles, original entered, readOnly readOnlyContent) (form, tea.Cmd) {
	filled := form{
		keys:        keys,
		title:       input.NewLine("", styles),
		description: input.NewLine("", styles),
		tags:        original.tags,
		language:    original.language,
		content:     input.NewContentArea(styles),
		readOnly:    readOnly,
		original:    original,
		field:       domain.FieldTitle,
		inContent:   false,
		invalid:     nil,
	}
	filled.title.SetValue(original.title)
	filled.description.SetValue(original.description)

	return filled.withContent(original.content).focused(domain.FieldTitle)
}

func (f form) withContent(content string) form {
	if !f.readOnly.held() {
		f.content.SetValue(content)
	}

	return f
}

func (f form) withExternalContent(content, codeStyle string) form {
	next := f
	next.readOnly = readOnlyIfUneditable(content, f.language, codeStyle)

	if next.readOnly.held() {
		next = next.leftContent()
	}

	return next.withContent(content)
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

func (f form) restyled(styles look.Styles) form {
	f.title = input.RestyledLine(f.title, styles)
	f.description = input.RestyledLine(f.description, styles)
	f.content = input.RestyledArea(f.content, styles)
	f.readOnly = f.readOnly.highlightedIn(styles.CodeStyle)

	return f
}

func (f form) withInvalid(fieldErrors []domain.FieldError) form {
	f.invalid = make([]domain.Field, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		f.invalid = append(f.invalid, fieldErr.Field)
	}

	return f
}

func (f form) withLanguage(language value.Language, codeStyle string) form {
	f.language = language
	f.readOnly = f.readOnly.inLanguage(language, codeStyle)

	return f
}

func (f form) withTags(chosen tagchoice.Chosen) form {
	f.tags = chosen

	return f
}

func (f form) changed() bool {
	return !f.entered().equal(f.original)
}

func (f form) entered() entered {
	content := f.content.Value()
	if f.readOnly.held() {
		content = f.readOnly.content
	}

	return entered{
		title:       f.title.Value(),
		description: f.description.Value(),
		tags:        f.tags,
		language:    f.language,
		content:     content,
	}
}

func (f form) hints() []key.Binding {
	if f.inContent {
		return f.keys.content.ShortHelp()
	}

	return f.keys.fields.ShortHelp()
}

func (f form) externalEditorKey() string {
	return f.keys.externalEditorKey()
}

func (f form) view(styles look.Styles, outer look.Size) string {
	lines := []string{
		f.fieldLine(styles, domain.FieldTitle, f.title.View()),
		f.fieldLine(styles, domain.FieldDescription, f.description.View()),
		f.fieldLine(styles, domain.FieldTagName, f.tagsEntry(styles)),
		f.fieldLine(styles, domain.FieldLanguage, f.languageEntry(styles)),
		f.fieldLine(styles, domain.FieldContent, f.contentEntryHint(styles)),
		f.contentView(),
	}

	return look.Frame(styles.Focused, f.frameTitle(), strings.Join(lines, "\n"), outer)
}

func (f form) contentView() string {
	if f.readOnly.held() {
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
	case fields.Matches(msg, binding.OpenInEditor):
		return f, requestExternalEditor, nil
	case fields.Matches(msg, binding.EditTags),
		f.field == domain.FieldTagName && fields.Matches(msg, binding.OpenField):
		return f, requestEditTags, nil
	case fields.Matches(msg, binding.PickLanguage),
		f.field == domain.FieldLanguage && fields.Matches(msg, binding.OpenField):
		return f, requestPickLanguage, nil
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
	case f.keys.content.Matches(msg, binding.EditTags):
		return f, requestEditTags, nil
	case f.keys.content.Matches(msg, binding.PickLanguage):
		return f, requestPickLanguage, nil
	case f.keys.content.Matches(msg, binding.OpenInEditor):
		return f, requestExternalEditor, nil
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

func overflowsEmptyContent(text string) bool {
	return emptyContentLines+lineBreaksOnInsert(text) > maxContentLines
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
		if f.readOnly.held() {
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
	if f.readOnly.held() {
		return styles.Dim.Render(f.readOnly.text(f.externalEditorKey()))
	}

	if f.inContent {
		return ""
	}

	entryKeys := f.firstKeysOf(binding.OpenField, binding.NextField)

	if len(entryKeys) == 0 {
		return ""
	}

	return styles.Dim.Render(strings.Join(entryKeys, entryKeysJoiner) + entrySuffix)
}

func (f form) tagsEntry(styles look.Styles) string {
	names := f.tags.Names()
	for index, name := range names {
		names[index] = tagPrefix + name
	}

	return f.withKeysHint(styles, strings.Join(names, tagSeparator), entrySuffix, binding.EditTags)
}

func (f form) languageEntry(styles look.Styles) string {
	return f.withKeysHint(styles, f.language.String(), pickSuffix, binding.PickLanguage)
}

func (f form) withKeysHint(styles look.Styles, shown, suffix, opener string) string {
	openKeys := f.firstKeysOf(binding.OpenField, opener)

	if len(openKeys) == 0 {
		return shown
	}

	return shown + styles.Dim.Render(keysHintOpen+strings.Join(openKeys, entryKeysJoiner)+suffix+keysHintClose)
}

func (f form) firstKeysOf(names ...string) []string {
	firstKeys := make([]string, 0, len(names))
	for _, name := range names {
		firstKeys = append(firstKeys, f.keys.fields.FirstKey(name))
	}

	return binding.BoundOnly(firstKeys...)
}

func (f form) frameTitle() string {
	if f.changed() {
		return overlayTitle + unsavedMarker
	}

	return overlayTitle
}

func editFields() []domain.Field {
	return []domain.Field{
		domain.FieldTitle, domain.FieldDescription, domain.FieldTagName, domain.FieldLanguage, domain.FieldContent,
	}
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
	case domain.FieldTagName:
		return "Tags"
	case domain.FieldFolderName:
	}

	return field.String()
}
