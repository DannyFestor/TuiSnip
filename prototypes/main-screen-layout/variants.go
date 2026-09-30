// PROTOTYPE — throwaway. Three structurally different layouts, each with its own navigation model.
package main

import (
	"charm.land/lipgloss/v2"
)

func maximizedPane(m *model, w, h int) string {
	switch m.focus {
	case paneSidebar:
		return box("Folders & Tags", m.sidebarRows(w-2, -1, true), w, h, true)
	case paneList:
		return box(m.listTitle(), m.listRows(w-2, h-3), w, h, true)
	}
	return box(m.snippetPaneTitle(), m.snippetPaneBody(w-2, h-3), w, h, true)
}

func drill(m *model, forward bool) {
	if forward && m.focus < paneSnippet {
		m.focus++
	}
	if !forward && m.focus > paneSidebar {
		m.focus--
	}
}

// A — classic three columns, one sidebar cursor across Folders and Tags, tab cycles panes.
type threeColumns struct{}

func (threeColumns) name() string          { return "Three columns · tab cycles panes" }
func (threeColumns) searchAsPalette() bool { return false }
func (threeColumns) sidebarSections() bool { return false }

func (threeColumns) navHelp() []string {
	return []string{"tab next pane", "shift+tab prev pane", "enter open (→)", "j/k ↑/↓ move", "g/G top/bottom", "space collapse Folder"}
}

func (threeColumns) nav(m *model, key string) bool {
	switch key {
	case "tab":
		m.focus = (m.focus + 1) % 3
	case "shift+tab":
		m.focus = (m.focus + 2) % 3
	case "enter":
		drill(m, true)
	default:
		return m.moveCursor(key)
	}
	return true
}

func (threeColumns) layout(m *model, w, h int) string {
	if m.maximized {
		return maximizedPane(m, w, h)
	}
	sideW, listW := 26, 36
	paneW := w - sideW - listW
	return lipgloss.JoinHorizontal(lipgloss.Top,
		box("Folders", m.sidebarRows(sideW-2, -1, true), sideW, h, m.focus == paneSidebar),
		box(m.listTitle(), m.listRows(listW-2, h-3), listW, h, m.focus == paneList),
		box(m.snippetPaneTitle(), m.snippetPaneBody(paneW-2, h-3), paneW, h, m.focus == paneSnippet),
	)
}

// B — mail-client: sidebar tabs (Folders | Tags), list stacked above the Snippet pane, h/l hop panes.
type stacked struct{}

func (stacked) name() string          { return "Sidebar tabs · list above preview · h/l hop" }
func (stacked) searchAsPalette() bool { return false }
func (stacked) sidebarSections() bool { return true }

func (stacked) navHelp() []string {
	return []string{"h/← prev pane", "l/→ next pane", "enter open (→)", "tab Folders ⇄ Tags (sidebar)", "j/k ↑/↓ move", "g/G top/bottom"}
}

func (stacked) nav(m *model, key string) bool {
	switch key {
	case "h", "left":
		drill(m, false)
	case "l", "right", "enter":
		drill(m, true)
	case "tab":
		if m.focus != paneSidebar {
			return false
		}
		m.setSection(entryFolder + entryTag - m.sideSection)
		m.browseSection = m.sideSection
	default:
		return m.moveCursor(key)
	}
	return true
}

func (stacked) layout(m *model, w, h int) string {
	if m.maximized {
		return maximizedPane(m, w, h)
	}
	sideW := 28
	rightW := w - sideW
	listH := max(6, h*2/5)
	tabs := "[Folders]  Tags "
	if m.sideSection == entryTag {
		tabs = " Folders  [Tags]"
	}
	right := lipgloss.JoinVertical(lipgloss.Left,
		box(m.listTitle(), m.listRows(rightW-2, listH-3), rightW, listH, m.focus == paneList),
		box(m.snippetPaneTitle(), m.snippetPaneBody(rightW-2, h-listH-3), rightW, h-listH, m.focus == paneSnippet),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		box(tabs, m.sidebarRows(sideW-2, m.sideSection, true), sideW, h, m.focus == paneSidebar),
		right,
	)
}

// C — the focused pane widens; enter drills in, esc backs out, 1/2/3 jump; Search is a palette.
type focusWidens struct{}

func (focusWidens) name() string          { return "Focus widens · enter in, esc out · Search palette" }
func (focusWidens) searchAsPalette() bool { return true }
func (focusWidens) sidebarSections() bool { return true }

func (focusWidens) navHelp() []string {
	return []string{"enter/l drill in (→)", "esc/h back out (←)", "1 2 3 jump to pane", "tab Folders ⇄ Tags (sidebar)", "j/k ↑/↓ move", "/ palette: type, ↑/↓, enter reveals"}
}

func (focusWidens) nav(m *model, key string) bool {
	switch key {
	case "enter", "l", "right":
		drill(m, true)
	case "esc", "h", "left", "backspace":
		drill(m, false)
	case "1", "2", "3":
		m.focus = pane(key[0] - '1')
	case "tab":
		if m.focus != paneSidebar {
			return false
		}
		m.setSection(entryFolder + entryTag - m.sideSection)
		m.browseSection = m.sideSection
	default:
		return m.moveCursor(key)
	}
	return true
}

func (focusWidens) widths(m *model, w int) (side, list int) {
	switch m.focus {
	case paneSidebar:
		return 34, 40
	case paneList:
		return 18, 52
	}
	return 18, 30
}

func (v focusWidens) layout(m *model, w, h int) string {
	var base string
	if m.maximized {
		base = maximizedPane(m, w, h)
	} else {
		sideW, listW := v.widths(m, w)
		paneW := w - sideW - listW
		tagsH := max(5, h/3)
		folders := box("1 Folders", m.sidebarRows(sideW-2, entryFolder, m.sideSection == entryFolder), sideW, h-tagsH, m.focus == paneSidebar && m.sideSection == entryFolder)
		tags := box("Tags", m.sidebarRows(sideW-2, entryTag, m.sideSection == entryTag), sideW, tagsH, m.focus == paneSidebar && m.sideSection == entryTag)
		base = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Left, folders, tags),
			box("2 "+m.listTitle(), m.listRows(listW-2, h-3), listW, h, m.focus == paneList),
			box("3 "+m.snippetPaneTitle(), m.snippetPaneBody(paneW-2, h-3), paneW, h, m.focus == paneSnippet),
		)
	}
	if !m.searching {
		return base
	}
	return v.withPalette(m, base, w, h)
}

func (focusWidens) withPalette(m *model, base string, w, h int) string {
	pw, ph := w*4/5, h*4/5
	resultsW := pw * 2 / 5
	results := m.searchLine(resultsW-2) + "\n\n" + m.listRows(resultsW-2, ph-6)
	preview := m.snippetPaneBody(pw-resultsW-2, ph-3)
	palette := lipgloss.JoinHorizontal(lipgloss.Top,
		box("Search", results, resultsW, ph, true),
		box("Preview", preview, pw-resultsW, ph, false),
	)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(palette).X((w-pw)/2).Y((h-ph)/2).Z(1),
	).Render()
}

// D — revision after feedback: Folders and Tags stacked as panes 1 and 2, list 3, Snippet pane 4.
type fourPanes struct{}

const (
	stopFolders = iota
	stopTags
	stopList
	stopSnippet
	stopCount
)

func (fourPanes) name() string          { return "Four panes · 1 Folders 2 Tags 3 list 4 Snippet" }
func (fourPanes) searchAsPalette() bool { return false }
func (fourPanes) sidebarSections() bool { return true }
func (fourPanes) popupSearch()          {}

func (fourPanes) navHelp() []string {
	return []string{"tab / shift+tab cycle all four panes", "l/→ right · h/← left (Folders and Tags are one column)", "1 2 3 4 jump to pane", "enter drill in (Folders/Tags → list → Snippet)", "esc back out", "j/k ↑/↓ move within pane", "/ Search popup"}
}

func stop(m *model) int {
	if m.focus == paneSidebar {
		if m.sideSection == entryTag {
			return stopTags
		}
		return stopFolders
	}
	return int(m.focus) + 1
}

func setStop(m *model, s int) {
	switch s {
	case stopFolders:
		m.focus = paneSidebar
		m.setSection(entryFolder)
	case stopTags:
		m.focus = paneSidebar
		m.setSection(entryTag)
	default:
		m.focus = pane(s - 1)
	}
}

func (fourPanes) nav(m *model, key string) bool {
	s := stop(m)
	switch key {
	case "tab":
		setStop(m, (s+1)%stopCount)
	case "shift+tab":
		setStop(m, (s+stopCount-1)%stopCount)
	case "l", "right":
		stepRight(m, s)
	case "h", "left":
		stepLeft(m, s)
	case "1", "2", "3", "4":
		setStop(m, int(key[0]-'1'))
	case "enter":
		if m.focus == paneSidebar {
			m.browseSection = m.sideSection
		}
		setStop(m, min(max(s+1, stopList), stopSnippet))
	case "esc":
		if s >= stopList {
			stepLeft(m, s)
		}
	default:
		return m.moveCursor(key)
	}
	return true
}

// stepRight and stepLeft treat Folders and Tags as one column, so sideways moves skip Tags.
func stepRight(m *model, s int) {
	setStop(m, min(max(s+1, stopList), stopSnippet))
}

func stepLeft(m *model, s int) {
	switch s {
	case stopSnippet:
		setStop(m, stopList)
	case stopList:
		m.focus = paneSidebar
		m.setSection(m.browseSection)
	}
}

func (fourPanes) widths(m *model) (side, list int) {
	switch m.focus {
	case paneSidebar:
		return 32, 34
	case paneList:
		return 24, 44
	}
	return 22, 32
}

// sidebarHeights makes the active left pane tall: the focused one, else the one filling the list.
func (fourPanes) sidebarHeights(m *model, h int) (folders, tags int) {
	tall := m.browseSection
	if m.focus == paneSidebar {
		tall = m.sideSection
	}
	big := h * 2 / 3
	if tall == entryTag {
		return h - big, big
	}
	return big, h - big
}

func (v fourPanes) layout(m *model, w, h int) string {
	if m.maximized {
		return maximizedPane(m, w, h)
	}
	sideW, listW := v.widths(m)
	paneW := w - sideW - listW
	foldersH, tagsH := v.sidebarHeights(m, h)
	s := stop(m)
	folders := box("1 Folders", m.sidebarRows(sideW-2, entryFolder, true), sideW, foldersH, s == stopFolders)
	tags := box("2 Tags", m.sidebarRows(sideW-2, entryTag, true), sideW, tagsH, s == stopTags)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, folders, tags),
		box("3 "+m.listTitle(), m.listRows(listW-2, h-3), listW, h, s == stopList),
		box("4 "+m.snippetPaneTitle(), m.snippetPaneBody(paneW-2, h-3), paneW, h, s == stopSnippet),
	)
}
