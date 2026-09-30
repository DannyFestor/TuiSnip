// PROTOTYPE — throwaway. Shared state and key routing for every variant.
package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type pane int

const (
	paneSidebar pane = iota
	paneList
	paneSnippet
)

var paneNames = [...]string{"sidebar", "snippet_list", "snippet_pane"}

type entryKind int

const (
	entryRoot entryKind = iota
	entryFolder
	entryTag
)

type entry struct {
	kind  entryKind
	id    string
	depth int
}

type sortOrder int

const (
	sortTitle sortOrder = iota
	sortUpdated
	sortCreated
)

var sortNames = [...]string{"title", "updated", "created"}

// variant is one layout plus its navigation model; everything else is shared.
type variant interface {
	name() string
	navHelp() []string
	// nav handles navigation keys for the pane Scopes; true means consumed.
	nav(m *model, key string) bool
	layout(m *model, width, height int) string
	// searchAsPalette lets a variant render Search as an overlay instead of in the list.
	searchAsPalette() bool
	// sidebarSections reports whether the Folder tree and Tag list are separate cursor stops.
	sidebarSections() bool
}

// popupSearcher marks a variant whose Search is a centred popup rather than a mode of the list.
type popupSearcher interface{ popupSearch() }

type model struct {
	store    *store
	variants []variant
	current  int

	width, height int
	focus         pane
	maximized     bool
	wrap          bool
	sort          sortOrder

	sideCursor int
	// otherSideCursor keeps the inactive section's place so switching Folders ⇄ Tags doesn't lose it.
	otherSideCursor int
	// browseSection is the sidebar section whose selection fills the list; focus alone never changes it.
	browseSection entryKind
	sideSection   entryKind // entryFolder or entryTag when sections are separate
	listCursor    int
	scroll        viewport.Model

	searching     bool
	searchFocused bool
	search        textinput.Model

	inlineEdit   *inlineEdit
	editor       *editor
	overlay      overlay
	status       string
	quitting     bool
	lastSelected string
	editPreview  viewport.Model
	editShown    string
}

type inlineEdit struct {
	input  textinput.Model
	target entry
	isNew  bool
}

func newModel(s *store, variants []variant, start int) *model {
	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "Search title, Tags, Description, content"
	m := &model{store: s, variants: variants, current: start, search: search, scroll: viewport.New(), editPreview: viewport.New(), sideSection: entryFolder, browseSection: entryFolder}
	m.status = "PROTOTYPE: nothing is saved. ` / ~ switch variant, ? help."
	return m
}

func (m *model) variant() variant { return m.variants[m.current] }

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.PasteMsg:
		return m, m.routePaste(msg)
	case tea.KeyPressMsg:
		return m, m.routeKey(msg)
	}
	return m, nil
}

func (m *model) routePaste(msg tea.PasteMsg) tea.Cmd {
	if m.editor != nil && m.overlay == nil {
		return m.editor.update(m, msg)
	}
	return nil
}

func (m *model) routeKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" {
		return m.confirmIfDirty("Quit with unsaved changes?", func() tea.Cmd { return tea.Quit })
	}
	if m.overlay != nil {
		return m.overlay.update(m, msg)
	}
	if m.inlineEdit != nil {
		return m.updateInlineEdit(msg)
	}
	if m.editor != nil && m.focus == paneSnippet {
		return m.editor.update(m, msg)
	}
	if m.searching && m.searchFocused {
		return m.updateSearch(msg)
	}
	if m.switchVariant(key) || m.global(key) || m.variant().nav(m, key) {
		if m.quitting {
			return tea.Quit
		}
		return nil
	}
	m.paneBinding(key)
	return nil
}

func (m *model) switchVariant(key string) bool {
	switch key {
	case "`":
		m.current = (m.current + 1) % len(m.variants)
	case "~":
		m.current = (m.current + len(m.variants) - 1) % len(m.variants)
	default:
		return false
	}
	m.closeSearch()
	m.status = "Variant " + m.variant().name()
	return true
}

func (m *model) global(key string) bool {
	switch key {
	case "q":
		m.quitting = true
	case "?":
		m.overlay = &helpOverlay{}
	case "/":
		if _, ok := m.variant().(popupSearcher); ok {
			m.overlay = newSearchPopup()
			return true
		}
		m.openSearch()
	case "z":
		m.maximized = !m.maximized
	case "n":
		m.startEditor(m.newSnippet(""))
	case "p":
		m.startEditor(m.newSnippet("kubectl get pods -A --field-selector=status.phase!=Running\n"))
		m.status = "Captured (prototype fakes the clipboard read)"
	default:
		return false
	}
	return true
}

func (m *model) paneBinding(key string) {
	switch m.focus {
	case paneSidebar:
		m.sidebarBinding(key)
	case paneList:
		m.listBinding(key)
	case paneSnippet:
		m.snippetPaneBinding(key)
	}
}

func (m *model) sidebarBinding(key string) {
	e, ok := m.selectedEntry()
	if !ok {
		return
	}
	switch key {
	case "N":
		m.inlineEdit = newInlineEdit(e, true, "")
	case "r":
		if e.kind != entryRoot {
			m.inlineEdit = newInlineEdit(e, false, m.entryName(e))
		}
	case "d":
		m.confirmDeleteEntry(e)
	case "m":
		if e.kind == entryFolder {
			m.overlay = newFolderPicker(m, "Move Folder "+m.entryName(e)+" to", e.id, func(dest string) {
				m.store.folder(e.id).parent = dest
				m.status = "Moved Folder to " + m.store.path(dest)
			})
		}
	case "space":
		if e.kind == entryFolder {
			f := m.store.folder(e.id)
			f.collapsed = !f.collapsed
		}
	}
}

func (m *model) listBinding(key string) {
	sn := m.selectedSnippet()
	if sn == nil {
		return
	}
	switch key {
	case "y":
		m.status = "Copied (prototype: not really)"
	case "e":
		m.startEditor(sn)
	case "E":
		m.status = "Would suspend and open " + sn.lang + " content in $EDITOR"
	case "m":
		m.overlay = newFolderPicker(m, "Move Snippet "+sn.title+" to", "", func(dest string) {
			sn.folder = dest
			m.status = "Moved Snippet to " + m.store.path(dest)
		})
	case "c":
		dup := *sn
		dup.id, dup.tags = m.store.newID(), slices.Clone(sn.tags)
		m.store.snippets = append(m.store.snippets, &dup)
		m.status = "Duplicated"
	case "d":
		m.overlay = newConfirm(fmt.Sprintf("Delete Snippet %q? [y/N]", sn.title), func() tea.Cmd {
			m.store.deleteSnippet(sn.id)
			m.status = "Deleted"
			return nil
		})
	case "s":
		m.sort = (m.sort + 1) % 3
		m.status = "Sorted by " + sortNames[m.sort]
	}
}

func (m *model) snippetPaneBinding(key string) {
	switch key {
	case "y":
		m.status = "Copied (prototype: not really)"
	case "e":
		if sn := m.selectedSnippet(); sn != nil {
			m.startEditor(sn)
		}
	case "E":
		m.status = "Would suspend and open the content in $EDITOR"
	case "w":
		m.wrap = !m.wrap
	case "j", "down":
		m.scroll.ScrollDown(1)
	case "k", "up":
		m.scroll.ScrollUp(1)
	case "pgdown", "ctrl+d":
		m.scroll.HalfPageDown()
	case "pgup", "ctrl+u":
		m.scroll.HalfPageUp()
	case "g", "home":
		m.scroll.GotoTop()
	case "G", "end":
		m.scroll.GotoBottom()
	}
}

// moveCursor is the shared up/down/top/bottom for the sidebar and the list.
func (m *model) moveCursor(key string) bool {
	delta := map[string]int{"j": 1, "down": 1, "k": -1, "up": -1, "pgdown": 10, "pgup": -10, "g": -1 << 20, "home": -1 << 20, "G": 1 << 20, "end": 1 << 20}[key]
	if delta == 0 {
		return false
	}
	switch m.focus {
	case paneSidebar:
		m.sideCursor = clamp(m.sideCursor+delta, 0, len(m.visibleEntries())-1)
		m.browseSection = m.sideSection
		m.listCursor = 0
	case paneList:
		m.listCursor = clamp(m.listCursor+delta, 0, len(m.listed())-1)
	case paneSnippet:
		m.snippetPaneBinding(key)
	}
	return true
}

func clamp(v, lo, hi int) int { return max(lo, min(v, hi)) }

func (m *model) entries() []entry {
	out := []entry{{kind: entryRoot}}
	var walk func(parent string, depth int)
	walk = func(parent string, depth int) {
		for _, f := range m.store.children(parent) {
			out = append(out, entry{kind: entryFolder, id: f.id, depth: depth})
			if !f.collapsed {
				walk(f.id, depth+1)
			}
		}
	}
	walk("", 1)
	for _, t := range m.store.tags {
		out = append(out, entry{kind: entryTag, id: t})
	}
	return out
}

// visibleEntries is what the sidebar cursor walks: all entries, or one section.
func (m *model) visibleEntries() []entry {
	all := m.entries()
	if !m.variant().sidebarSections() {
		return all
	}
	return slices.DeleteFunc(all, func(e entry) bool { return (e.kind == entryTag) != (m.sideSection == entryTag) })
}

func (m *model) selectedEntry() (entry, bool) {
	es := m.visibleEntries()
	if len(es) == 0 {
		return entry{}, false
	}
	m.sideCursor = clamp(m.sideCursor, 0, len(es)-1)
	return es[m.sideCursor], true
}

func (m *model) sectionEntries(section entryKind) []entry {
	return slices.DeleteFunc(m.entries(), func(e entry) bool { return (e.kind == entryTag) != (section == entryTag) })
}

// entryAt is the cursor's entry in either sidebar section, focused or not.
func (m *model) entryAt(section entryKind) (entry, bool) {
	if section == m.sideSection {
		return m.selectedEntry()
	}
	es := m.sectionEntries(section)
	if len(es) == 0 {
		return entry{}, false
	}
	m.otherSideCursor = clamp(m.otherSideCursor, 0, len(es)-1)
	return es[m.otherSideCursor], true
}

// browseEntry is the Folder, Root, or Tag whose Snippets the list shows.
func (m *model) browseEntry() (entry, bool) {
	if !m.variant().sidebarSections() {
		return m.selectedEntry()
	}
	return m.entryAt(m.browseSection)
}

func (m *model) entryName(e entry) string {
	switch e.kind {
	case entryRoot:
		return "Root"
	case entryFolder:
		return m.store.folder(e.id).name
	default:
		return e.id
	}
}

func (m *model) entryCount(e entry) int {
	switch e.kind {
	case entryRoot:
		return m.store.folderCount("")
	case entryFolder:
		return m.store.folderCount(e.id)
	default:
		return m.store.tagCount(e.id)
	}
}

// browseFolder is the Folder new Snippets land in: the selected Folder, else the Root.
func (m *model) browseFolder() string {
	if e, ok := m.browseEntry(); ok && e.kind == entryFolder {
		return e.id
	}
	return ""
}

func (m *model) listed() []*snippet {
	if m.searching && m.search.Value() != "" {
		return m.store.search(m.search.Value())
	}
	e, ok := m.browseEntry()
	if !ok {
		return nil
	}
	var out []*snippet
	switch e.kind {
	case entryRoot:
		out = m.store.inFolder("")
	case entryFolder:
		out = m.store.inFolder(e.id)
	default:
		out = m.store.withTag(e.id)
	}
	slices.SortStableFunc(out, m.compare)
	return out
}

func (m *model) compare(a, b *snippet) int {
	switch m.sort {
	case sortUpdated:
		return b.updated.Compare(a.updated)
	case sortCreated:
		return b.created.Compare(a.created)
	default:
		return strings.Compare(strings.ToLower(a.title), strings.ToLower(b.title))
	}
}

func (m *model) selectedSnippet() *snippet {
	if m.editor != nil {
		return m.editor.target
	}
	list := m.listed()
	if len(list) == 0 {
		return nil
	}
	m.listCursor = clamp(m.listCursor, 0, len(list)-1)
	return list[m.listCursor]
}

func (m *model) openSearch() {
	m.searching, m.searchFocused = true, true
	m.listCursor = 0
	m.search.Focus()
}

func (m *model) closeSearch() {
	m.searching, m.searchFocused = false, false
	m.search.Reset()
	m.search.Blur()
}

func (m *model) updateSearch(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.closeSearch()
		return nil
	case "enter":
		m.acceptSearch()
		return nil
	case "up", "ctrl+p":
		m.listCursor = max(0, m.listCursor-1)
		return nil
	case "down", "ctrl+n":
		m.listCursor = clamp(m.listCursor+1, 0, max(0, len(m.listed())-1))
		return nil
	}
	var cmd tea.Cmd
	before := m.search.Value()
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != before {
		m.listCursor = 0
	}
	return cmd
}

func (m *model) acceptSearch() {
	m.searchFocused = false
	m.search.Blur()
	if m.variant().searchAsPalette() {
		m.revealSelected()
		return
	}
	m.focus = paneList
}

// revealSelected jumps Browse to the chosen result's Folder and closes Search.
func (m *model) revealSelected() {
	sn := m.selectedSnippet()
	m.closeSearch()
	if sn != nil {
		m.revealSnippet(sn)
	}
}

func (m *model) setSection(k entryKind) {
	if m.sideSection == k {
		return
	}
	m.sideSection = k
	m.sideCursor, m.otherSideCursor = m.otherSideCursor, m.sideCursor
	m.listCursor = 0
}

func (m *model) revealSnippet(sn *snippet) {
	m.setSection(entryFolder)
	m.browseSection = entryFolder
	for i, e := range m.visibleEntries() {
		if (sn.folder == "" && e.kind == entryRoot) || (e.kind == entryFolder && e.id == sn.folder) {
			m.sideCursor = i
		}
	}
	m.listCursor = slices.Index(m.listed(), sn)
	m.focus = paneSnippet
}

func newInlineEdit(target entry, isNew bool, value string) *inlineEdit {
	in := textinput.New()
	in.Prompt = ""
	in.SetValue(value)
	in.Focus()
	return &inlineEdit{input: in, target: target, isNew: isNew}
}

func (m *model) updateInlineEdit(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.inlineEdit = nil
		return nil
	case "enter":
		m.commitInlineEdit()
		return nil
	}
	var cmd tea.Cmd
	m.inlineEdit.input, cmd = m.inlineEdit.input.Update(msg)
	return cmd
}

func (m *model) commitInlineEdit() {
	ie := m.inlineEdit
	m.inlineEdit = nil
	name := strings.TrimSpace(ie.input.Value())
	if name == "" {
		m.status = "Name cannot be blank"
		return
	}
	if ie.isNew {
		parent := ""
		if ie.target.kind == entryFolder {
			parent = ie.target.id
		}
		lang := "plaintext"
		if p := m.store.folder(parent); p != nil {
			lang = p.lang
		}
		m.store.folders = append(m.store.folders, &folder{id: m.store.newID(), name: name, parent: parent, lang: lang})
		m.status = "Created Folder " + name
		return
	}
	switch ie.target.kind {
	case entryFolder:
		m.store.folder(ie.target.id).name = name
	case entryTag:
		m.renameTag(ie.target.id, name)
	}
	m.status = "Renamed to " + name
}

func (m *model) renameTag(from, to string) {
	for _, sn := range m.store.snippets {
		for i, t := range sn.tags {
			if t == from {
				sn.tags[i] = to
			}
		}
		sn.tags = slices.Compact(sn.tags)
	}
	m.store.deleteTag(from)
	m.store.ensureTags([]string{to})
}

func (m *model) confirmDeleteEntry(e entry) {
	switch e.kind {
	case entryFolder:
		subs, sns := m.store.subtreeCounts(e.id)
		msg := fmt.Sprintf("Permanently delete Folder %q, %d subfolder(s), and %d Snippet(s)? This cannot be undone. [y/N]", m.entryName(e), subs, sns)
		m.overlay = newConfirm(msg, func() tea.Cmd {
			m.store.deleteFolder(e.id)
			m.status = "Deleted Folder"
			return nil
		})
	case entryTag:
		if n := m.store.tagCount(e.id); n > 0 {
			m.overlay = newConfirm(fmt.Sprintf("Delete Tag %q from %d Snippet(s)? [y/N]", e.id, n), func() tea.Cmd {
				m.store.deleteTag(e.id)
				return nil
			})
			return
		}
		m.store.deleteTag(e.id)
		m.status = "Deleted unused Tag"
	}
}

func (m *model) newSnippet(content string) *snippet {
	folderID := m.browseFolder()
	lang := "plaintext"
	if f := m.store.folder(folderID); f != nil {
		lang = f.lang
	}
	now := time.Now()
	var tags []string
	if e, ok := m.browseEntry(); ok && e.kind == entryTag {
		tags = []string{e.id} // a new Snippet made while browsing a Tag lands at the Root carrying it
	}
	return &snippet{folder: folderID, lang: lang, content: content, tags: tags, created: now, updated: now}
}

func (m *model) startEditor(sn *snippet) {
	m.closeSearch()
	m.editor = newEditor(sn)
	m.focus = paneSnippet
}

func (m *model) confirmIfDirty(question string, then func() tea.Cmd) tea.Cmd {
	if m.editor != nil && m.editor.dirty() {
		m.overlay = newConfirm(question+" [y/N]", then)
		return nil
	}
	return then()
}

func (m *model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}
