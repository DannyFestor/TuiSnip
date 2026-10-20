package editoverlay

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/input"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagchoice"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	overlayTitle    = "Editing"
	unsavedMarker   = " •"
	fieldCursor     = "› "
	fieldIndent     = "  "
	fieldLabelWidth = 15
	fieldRows       = 6
	tagPrefix       = "#"
	tagSeparator    = " "
	entryKeysJoiner = " or "
	entrySuffix     = " to edit"
	pickSuffix      = " to pick"
	keysHintOpen    = "   ("
	keysHintClose   = ")"
)

type form struct {
	keys          formKeys
	title         textinput.Model
	description   textinput.Model
	tags          tagchoice.Chosen
	fragmentTitle textinput.Model
	language      value.Language
	content       textarea.Model
	tabs          fragmentTabs
	outer         look.Size
	original      entered
	field         domain.Field
	inContent     bool
	invalid       []domain.Field
}

func newForm(keys formKeys, styles look.Styles, original entered, shown int) (form, tea.Cmd) {
	filled := form{
		keys:          keys,
		title:         input.NewLine("", styles),
		description:   input.NewLine("", styles),
		tags:          original.tags,
		fragmentTitle: input.NewLine("", styles),
		language:      value.Language{},
		content:       input.NewContentArea(styles),
		tabs:          newFragmentTabs(styles, original.fragments, shown),
		outer:         look.Size{Width: 0, Height: 0},
		original:      original,
		field:         domain.FieldTitle,
		inContent:     false,
		invalid:       nil,
	}
	filled.title.SetValue(original.title)
	filled.description.SetValue(original.description)

	return filled.withDraft(filled.tabs.shownDraft()).focused(domain.FieldTitle)
}

func (f form) withContent(content string) form {
	f.content.SetValue(content)

	return f
}

func (f form) update(msg tea.Msg) (form, request, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if f.inContent {
			return f.contentPressed(msg)
		}

		if next, asked, ok := f.fragmentKeyPressed(msg); ok {
			return next, asked, nil
		}

		return f.fieldPressed(msg)
	case tea.PasteMsg:
		return f.pasted(msg)
	case pointer.Clicked:
		return f.clicked(msg)
	case pointer.ClickedOutside:
		return f.clickedOutside()
	}

	return f, requestNothing, nil
}

func (f form) resized(outer look.Size) form {
	inner := outer.Inner()
	inputWidth := max(1, inner.Width-len(fieldIndent)-fieldLabelWidth-input.CursorWidth)

	f.outer = outer
	f.title.SetWidth(inputWidth)
	f.description.SetWidth(inputWidth)
	f.fragmentTitle.SetWidth(inputWidth)
	f.content.SetWidth(inner.Width)
	f.content.SetHeight(max(1, inner.Height-fieldRows-f.stripRows()))

	return f
}

func (f form) restyled(styles look.Styles) form {
	f.title = input.RestyledLine(f.title, styles)
	f.description = input.RestyledLine(f.description, styles)
	f.fragmentTitle = input.RestyledLine(f.fragmentTitle, styles)
	f.content = input.RestyledArea(f.content, styles)
	f.tabs = f.tabs.restyled(styles)

	return f
}

func (f form) withInvalid(fieldErrors []domain.FieldError) form {
	f.invalid = make([]domain.Field, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		f.invalid = append(f.invalid, fieldErr.Field)
	}

	return f
}

func (f form) withLanguage(language value.Language) form {
	f.language = language

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
	return entered{
		title:       f.title.Value(),
		description: f.description.Value(),
		tags:        f.tags,
		fragments:   f.fragments().all(),
	}
}

func (f form) hints() []key.Binding {
	if f.inContent {
		return f.keys.content.ShortHelp()
	}

	return f.keys.fields.ShortHelp()
}

func (f form) view(styles look.Styles, outer look.Size) string {
	lines := []string{
		f.fieldLine(styles, domain.FieldTitle, f.title.View()),
		f.fieldLine(styles, domain.FieldDescription, f.description.View()),
		f.fieldLine(styles, domain.FieldTagName, f.tagsEntry(styles)),
	}

	if tabs := f.fragments(); tabs.visible() {
		lines = append(lines, fieldIndent+tabs.view(outer.Inner().Width-len(fieldIndent)))
	}

	lines = append(lines,
		f.fieldLine(styles, domain.FieldFragmentTitle, f.fragmentTitle.View()),
		f.fieldLine(styles, domain.FieldLanguage, f.languageEntry(styles)),
		f.fieldLine(styles, domain.FieldContent, f.contentEntryHint(styles)),
		f.content.View(),
	)

	return look.Frame(styles.Focused, f.frameTitle(), strings.Join(lines, "\n"), outer)
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

func (f form) clicked(click pointer.Clicked) (form, request, tea.Cmd) {
	point := click.InsideFrame().At
	line := point.Y

	if f.fragments().visible() {
		switch {
		case line == stripRow():
			return f.withFragmentClicked(point.X - len(fieldIndent)), requestNothing, nil
		case line > stripRow():
			line--
		}
	}

	return f.fieldClicked(line)
}

func (f form) fieldClicked(line int) (form, request, tea.Cmd) {
	fields := editFields()

	switch {
	case line < 0:
		return f, requestNothing, nil
	case line < len(fields):
		next, cmd := f.focused(fields[line])

		return next, requestNothing, cmd
	}

	return f.contentClicked()
}

func (f form) contentClicked() (form, request, tea.Cmd) {
	onContent, _ := f.focused(domain.FieldContent)
	next, cmd := onContent.enteredContent()

	return next, requestNothing, cmd
}

func (f form) clickedOutside() (form, request, tea.Cmd) {
	if f.inContent {
		return f.leftContent(), requestNothing, nil
	}

	return f, requestCancel, nil
}

func (f form) leavesContentUpward(msg tea.KeyPressMsg) bool {
	return f.content.Line() == 0 && key.Matches(msg, f.content.KeyMap.LinePrevious)
}

func (f form) pasted(msg tea.PasteMsg) (form, request, tea.Cmd) {
	if f.inContent && !value.ContentFits(f.bytesAfterPaste(msg.Content)) {
		return f, requestRefuseOversizedPaste, nil
	}

	return f.typed(msg)
}

func (f form) bytesAfterPaste(pasted string) int {
	return len(f.content.Value()) - len(f.content.SelectedText()) + len(pasted)
}

func (f form) typed(msg tea.Msg) (form, request, tea.Cmd) {
	var cmd tea.Cmd

	switch f.field {
	case domain.FieldTitle:
		f.title, cmd = f.title.Update(msg)
	case domain.FieldDescription:
		f.description, cmd = f.description.Update(msg)
	case domain.FieldFragmentTitle:
		f.fragmentTitle, cmd = f.fragmentTitle.Update(msg)
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
	f.fragmentTitle.Blur()

	switch field {
	case domain.FieldTitle:
		return f, f.title.Focus()
	case domain.FieldDescription:
		return f, f.description.Focus()
	case domain.FieldFragmentTitle:
		return f, f.fragmentTitle.Focus()
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
		domain.FieldTitle,
		domain.FieldDescription,
		domain.FieldTagName,
		domain.FieldFragmentTitle,
		domain.FieldLanguage,
		domain.FieldContent,
	}
}

func stripRow() int {
	return slices.Index(editFields(), domain.FieldFragmentTitle)
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
	case domain.FieldFragmentTitle:
		return "Fragment title"
	case domain.FieldFolderName:
	}

	return field.String()
}
