package textarea

import (
	"strings"

	"github.com/rivo/uniseg"

	"tuisnip-editor-prototype/internal/memoization"
)

// PROTOTYPE knobs, flipped live by the prototype app.
var (
	TabWidth = 4
	// TabStops: a tab runs to the next multiple of TabWidth. Off: every tab is TabWidth cells.
	TabStops = true
	// TabGlyph is drawn in a tab's first cell; empty draws plain spaces.
	TabGlyph = ""
	// CursorOnTab: "whole" covers every cell of the tab, "first" its first cell, "last" its last cell.
	CursorOnTab = "whole"
	// PerCellTabs: left/right step through a tab cell by cell instead of jumping over it.
	PerCellTabs = false
	// WindowedRender draws only the rows near the viewport; the rest become blank rows.
	WindowedRender = true
)

func tabCells(col int) int {
	if TabStops {
		return TabWidth - col%TabWidth
	}
	return TabWidth
}

// widthFrom is the cell width of seg when it starts at cell col of its row.
func widthFrom(seg []rune, col int) int {
	end := col
	chunkStart := 0
	for i, r := range seg {
		if r != '\t' {
			continue
		}
		end += uniseg.StringWidth(string(seg[chunkStart:i]))
		end += tabCells(end)
		chunkStart = i + 1
	}
	end += uniseg.StringWidth(string(seg[chunkStart:]))
	return end - col
}

// expandFrom renders seg as terminal-safe text, each tab replaced by the cells it covers.
func expandFrom(seg []rune, col int) string {
	var b strings.Builder
	chunkStart := 0
	for i, r := range seg {
		if r != '\t' {
			continue
		}
		chunk := string(seg[chunkStart:i])
		b.WriteString(chunk)
		col += uniseg.StringWidth(chunk)
		b.WriteString(tabFill(tabCells(col)))
		col += tabCells(col)
		chunkStart = i + 1
	}
	b.WriteString(string(seg[chunkStart:]))
	return b.String()
}

func tabFill(cells int) string {
	if TabGlyph == "" || cells == 0 {
		return strings.Repeat(" ", cells)
	}
	return TabGlyph + strings.Repeat(" ", cells-1)
}

func (m Model) onTab() bool {
	return m.col < len(m.value[m.row]) && m.value[m.row][m.col] == '\t'
}

func (m Model) cursorTabCells() int {
	return tabCells(m.LineInfo().CharOffset - m.tabCell)
}

// TabCell is the cell inside a tab the cursor sits on, for the prototype's status bar.
func (m Model) TabCell() int {
	return m.tabCell
}

// RuneUnderCursor is the rune the cursor sits on, or -1 at the end of a line.
func (m Model) RuneUnderCursor() rune {
	if m.col >= len(m.value[m.row]) {
		return -1
	}
	return m.value[m.row][m.col]
}

// DedentLine removes one leading tab, or up to TabWidth leading spaces, from the cursor's line.
func (m *Model) DedentLine() {
	line := m.value[m.row]
	removed := 0
	if len(line) > 0 && line[0] == '\t' {
		removed = 1
	} else {
		for removed < len(line) && removed < TabWidth && line[removed] == ' ' {
			removed++
		}
	}
	m.value[m.row] = line[removed:]
	m.SetCursorColumn(max(0, m.col-removed))
}

// ResetWrapCache drops wrapped rows measured with the previous tab knobs.
func (m *Model) ResetWrapCache() {
	m.cache = memoization.NewMemoCache[line, [][]rune](m.cache.Capacity())
}

// cursorOnTab splits a tab of the given cells into the text before the cursor cell,
// the cursor cell, and the text after it.
func cursorOnTab(cells, tabCell int) (before, under, after string) {
	fill := []rune(tabFill(cells))
	at := 0
	switch {
	case PerCellTabs:
		at = min(tabCell, cells-1)
	case CursorOnTab == "whole":
		return "", string(fill), ""
	case CursorOnTab == "last":
		at = cells - 1
	}
	return string(fill[:at]), string(fill[at]), string(fill[at+1:])
}
