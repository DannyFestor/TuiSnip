package textarea

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tabstop"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea/memoization"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/textarea/runeutil"
)

const (
	minHeight        = 1
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 0
	defaultMaxHeight = 99
	defaultMaxWidth  = 500

	defaultMaxLines = 10000

	// 2 columns to account for the trailing space wrapping.
	trailingSpace = 2

	lineNumberGap = 2
)

// line is the input to the text wrapping function. This is stored in a struct
// so that it can be hashed and memoized.
type line struct {
	runes []rune
	width int
}

func (w line) Hash() string {
	v := fmt.Sprintf("%s:%d", string(w.runes), w.width)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
}

type Model struct {
	cache *memoization.MemoCache[line, []wrappedLine]

	// When changing the value of Prompt after the model has been
	// initialized, ensure that SetWidth() gets called afterwards.
	Prompt string

	Placeholder          string
	ShowLineNumbers      bool
	EndOfBufferCharacter rune
	KeyMap               KeyMap

	keyActions    []keyAction
	virtualCursor cursor.Model

	// CharLimit is the maximum number of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	MaxLines int

	// MaxHeight is the maximum height of the text area in rows. If 0 or less,
	// there's no limit.
	MaxHeight int

	// MaxWidth is the maximum width of the text area in columns. If 0 or less,
	// there's no limit.
	MaxWidth int

	// DynamicHeight, when true, causes the textarea to automatically grow
	// and shrink its height to fit the content. The height is clamped between
	// MinHeight and MaxHeight.
	DynamicHeight bool

	// MinHeight is the minimum height of the text area in rows when
	// DynamicHeight is enabled. If 0 or less, defaults to 1.
	MinHeight int

	// MaxContentHeight is the maximum content height in visual rows
	// (accounting for soft wraps). When set (> 0), input is blocked once
	// the total visual lines reach this limit, while MaxHeight controls
	// only the visible viewport height. When 0, the content guard falls
	// back to the legacy MaxHeight behavior (blocking at MaxHeight
	// logical lines) for backward compatibility.
	MaxContentHeight int

	styles           Styles
	useVirtualCursor bool
	promptFunc       func(PromptInfo) string
	promptWidth      int
	width            int
	height           int
	value            [][]rune
	focus            bool
	col              int
	row              int

	// Last character offset, used to maintain state when the cursor is moved
	// vertically such that we can maintain the same navigating position.
	lastCharOffset int

	viewport *viewport.Model
	rsan     *runeutil.Sanitizer

	// Selection anchor and head. The anchor is where the selection began;
	// the head follows the pointer. Either may sort before the other — use
	// [Model.Selection] for a normalized range.
	selAnchor Position
	selHead   Position

	// hasSelection reports whether a selection has been set at all.
	// [Model.HasSelection] additionally requires it to be non-empty.
	hasSelection bool

	selecting bool
}

func New() Model {
	scroller := viewport.New()

	var noKeys viewport.KeyMap

	scroller.KeyMap = noKeys

	model := Model{
		CharLimit:            defaultCharLimit,
		MaxLines:             defaultMaxLines,
		MaxHeight:            defaultMaxHeight,
		MaxWidth:             defaultMaxWidth,
		Prompt:               lipgloss.ThickBorder().Left + " ",
		Placeholder:          "",
		styles:               DefaultDarkStyles(),
		cache:                memoization.NewMemoCache[line, []wrappedLine](defaultMaxLines),
		EndOfBufferCharacter: ' ',
		ShowLineNumbers:      true,
		useVirtualCursor:     true,
		virtualCursor:        cursor.New(),
		KeyMap:               DefaultKeyMap(),
		keyActions:           newKeyActions(),
		DynamicHeight:        false,
		MinHeight:            0,
		MaxContentHeight:     0,
		promptFunc:           nil,
		promptWidth:          0,
		width:                0,
		height:               0,
		lastCharOffset:       0,
		rsan:                 nil,
		selAnchor:            Position{Row: 0, Col: 0},
		selHead:              Position{Row: 0, Col: 0},
		hasSelection:         false,
		selecting:            false,

		value: make([][]rune, minHeight, defaultMaxLines),
		focus: false,
		col:   0,
		row:   0,

		viewport: &scroller,
	}

	model.SetHeight(defaultHeight)
	model.SetWidth(defaultWidth)

	return model
}

func (m *Model) Styles() Styles {
	return m.styles
}

func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.updateVirtualCursorStyle()
}

func (m *Model) VirtualCursor() bool {
	return m.useVirtualCursor
}

func (m *Model) SetVirtualCursor(v bool) {
	m.useVirtualCursor = v
	m.updateVirtualCursorStyle()
}

func (m *Model) SetValue(s string) {
	m.Reset()
	m.InsertString(s)
	m.recalculateHeight()
}

func (m *Model) InsertString(s string) {
	m.insertRunesFromUserInput([]rune(s))
	m.recalculateHeight()
}

func (m *Model) InsertRune(r rune) {
	m.insertRunesFromUserInput([]rune{r})
	m.recalculateHeight()
}

func (m *Model) Value() string {
	if m.value == nil {
		return ""
	}

	var value strings.Builder
	for _, l := range m.value {
		value.WriteString(string(l))
		value.WriteByte('\n')
	}

	return strings.TrimSuffix(value.String(), "\n")
}

func (m *Model) Length() int {
	var l int
	for _, row := range m.value {
		l += uniseg.StringWidth(string(row))
	}
	// We add len(m.value) to include the newline characters.
	return l + len(m.value) - 1
}

func (m *Model) LineCount() int {
	return len(m.value)
}

func (m *Model) Line() int {
	return m.row
}

func (m *Model) Column() int {
	return m.col
}

func (m *Model) ScrollYOffset() int {
	return m.viewport.YOffset()
}

func (m *Model) ScrollPercent() float64 {
	return m.viewport.ScrollPercent()
}

func (m *Model) CursorDown() {
	m.setCursorLineRelative(1)
}

func (m *Model) CursorUp() {
	m.setCursorLineRelative(-1)
}

func (m *Model) SetCursorColumn(col int) {
	m.col = clamp(col, 0, len(m.value[m.row]))
	// Any time that we move the cursor horizontally we need to reset the last
	// offset so that the horizontal position when navigating is adjusted.
	m.lastCharOffset = 0
}

func (m *Model) CursorStart() {
	m.SetCursorColumn(0)
}

func (m *Model) CursorEnd() {
	m.SetCursorColumn(len(m.value[m.row]))
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() tea.Cmd {
	m.focus = true

	return m.virtualCursor.Focus()
}

func (m *Model) Blur() {
	m.focus = false
	m.virtualCursor.Blur()
}

func (m *Model) Reset() {
	m.value = make([][]rune, minHeight, defaultMaxLines)
	m.col = 0
	m.row = 0
	m.viewport.GotoTop()
	m.SetCursorColumn(0)
	m.recalculateHeight()
	m.ClearSelection()
}

func (m *Model) Word() string {
	line := m.value[m.row]
	col := m.col - 1

	if col < 0 {
		return ""
	}

	if col >= len(line) {
		return ""
	}

	if unicode.IsSpace(line[col]) {
		return ""
	}

	start := col
	for start > 0 && !unicode.IsSpace(line[start-1]) {
		start--
	}

	end := col
	for end < len(line) && !unicode.IsSpace(line[end]) {
		end++
	}

	return string(line[start:end])
}

func (m *Model) LineInfo() LineInfo {
	grid := m.memoizedWrap(m.value[m.row], m.width)

	var counter int
	for segment, line := range grid {
		if counter+len(line.runes) == m.col && segment+1 < len(grid) {
			// We wrap around to the next line if we are at the end of the
			// previous line so that we can be at the very beginning of the row
			return LineInfo{
				CharOffset:      0,
				ColumnOffset:    0,
				Height:          len(grid),
				RowOffset:       segment + 1,
				StartColumn:     m.col,
				StartCharOffset: grid[segment+1].startCell,
				Width:           len(grid[segment+1].runes),
				CharWidth:       line.width(),
			}
		}

		if counter+len(line.runes) >= m.col {
			return LineInfo{
				CharOffset:      line.widthBefore(max(0, m.col-counter)),
				ColumnOffset:    m.col - counter,
				Height:          len(grid),
				RowOffset:       segment,
				StartColumn:     counter,
				StartCharOffset: line.startCell,
				Width:           len(line.runes),
				CharWidth:       line.width(),
			}
		}

		counter += len(line.runes)
	}

	return LineInfo{}
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) MoveToBegin() {
	m.row = 0
	m.SetCursorColumn(0)
	m.repositionView()
}

func (m *Model) MoveToEnd() {
	m.row = len(m.value) - 1
	m.SetCursorColumn(len(m.value[m.row]))
	m.repositionView()
}

// PageUp moves the cursor up by one page. First call snaps to the first visible
// line, subsequent calls move up by a full page.
func (m *Model) PageUp() {
	if offset := m.viewport.YOffset() - m.cursorLineNumber(); offset < 0 {
		m.setCursorLineRelative(offset)

		return
	}

	m.setCursorLineRelative(-m.height)
}

// PageDown moves the cursor down by one page. First call snaps to the last
// visible line, subsequent calls move down by a full page.
func (m *Model) PageDown() {
	if offset := m.cursorLineNumber() - m.viewport.YOffset(); offset < m.height-1 {
		m.setCursorLineRelative(m.height - 1 - offset)

		return
	}

	m.setCursorLineRelative(m.height)
}

// SetWidth sets the width of the textarea to fit exactly within the given width.
// This means that the textarea will account for the width of the prompt and
// whether or not line numbers are being shown.
//
// Ensure that SetWidth is called after setting the Prompt and ShowLineNumbers,
// It is important that the width of the textarea be exactly the given width
// and no more.
func (m *Model) SetWidth(width int) {
	// Update prompt width only if there is no prompt function as
	// [SetPromptFunc] updates the prompt width when it is called.
	if m.promptFunc == nil {
		m.promptWidth = uniseg.StringWidth(m.Prompt)
	}

	reservedOuter := m.activeStyle().Base.GetHorizontalFrameSize()

	reservedInner := m.promptWidth
	if m.ShowLineNumbers {
		reservedInner += numDigits(m.MaxHeight) + lineNumberGap
	}

	// Input width must be at least one more than the reserved inner and outer
	// width. This gives us a minimum input width of 1.
	minWidth := reservedInner + reservedOuter + 1
	inputWidth := max(width, minWidth)

	if m.MaxWidth > 0 {
		inputWidth = min(inputWidth, m.MaxWidth)
	}

	// Since the width of the viewport and input area is dependent on the width of
	// borders, prompt and line numbers, we need to calculate it by subtracting
	// the reserved width from them.

	m.viewport.SetWidth(inputWidth - reservedOuter)
	m.width = inputWidth - reservedOuter - reservedInner
	m.recalculateHeight()
}

// SetPromptFunc supersedes the Prompt field and sets a dynamic prompt instead.
//
// If the function returns a prompt that is shorter than the specified
// promptWidth, it will be padded to the left. If it returns a prompt that is
// longer, display artifacts may occur; the caller is responsible for computing
// an adequate promptWidth.
func (m *Model) SetPromptFunc(promptWidth int, fn func(PromptInfo) string) {
	m.promptFunc = fn
	m.promptWidth = promptWidth
}

func (m *Model) Height() int {
	return m.height
}

func (m *Model) SetHeight(height int) {
	if m.MaxHeight > 0 {
		m.height = clamp(height, minHeight, m.MaxHeight)
		m.viewport.SetHeight(clamp(height, minHeight, m.MaxHeight))
	} else {
		m.height = max(height, minHeight)
		m.viewport.SetHeight(max(height, minHeight))
	}

	m.repositionView()
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.virtualCursor.Blur()

		return *m, nil
	}

	// Used to determine if the cursor should blink.
	oldRow, oldCol := m.cursorLineNumber(), m.col

	m.prepareForMsg()

	if m.handleMsg(msg).finished {
		return *m, nil
	}

	return *m, m.refresh(msg, oldRow, oldCol)
}

func (m *Model) View() string {
	// XXX: This is a workaround for the case where the viewport hasn't
	// been initialized yet like during the initial render. In that case,
	// we need to render the view again because Update hasn't been called
	// yet to set the content of the viewport.
	m.viewport.SetContent(m.contentView())
	view := m.viewport.View()
	styles := m.activeStyle()

	return styles.Base.Render(view)
}

// Cursor returns a [tea.Cursor] for rendering a real cursor in a Bubble Tea
// program. This requires that [Model.VirtualCursor] is set to false.
//
// Example:
//
//	// In your top-level View function:
//	f := tea.NewFrame(m.textarea.View())
//	f.Cursor = m.textarea.Cursor()
//	f.Cursor.Position.X += offsetX
//	f.Cursor.Position.Y += offsetY
func (m *Model) Cursor() *tea.Cursor {
	if m.useVirtualCursor || !m.Focused() {
		return nil
	}

	lineInfo := m.LineInfo()
	w := lipgloss.Width
	baseStyle := m.activeStyle().Base

	xOffset := lineInfo.CharOffset +
		w(m.promptView(0)) +
		w(m.lineNumberView(0)) +
		baseStyle.GetMarginLeft() +
		baseStyle.GetPaddingLeft() +
		baseStyle.GetBorderLeftSize()

	yOffset := m.cursorLineNumber() -
		m.viewport.YOffset() +
		baseStyle.GetMarginTop() +
		baseStyle.GetPaddingTop() +
		baseStyle.GetBorderTopSize()

	c := tea.NewCursor(xOffset, yOffset)
	c.Blink = m.styles.Cursor.Blink
	c.Color = m.styles.Cursor.Color
	c.Shape = m.styles.Cursor.Shape

	return c
}

func (m *Model) updateVirtualCursorStyle() {
	if !m.useVirtualCursor {
		m.virtualCursor.SetMode(cursor.CursorHide)

		return
	}

	m.virtualCursor.Style = lipgloss.NewStyle().Foreground(m.styles.Cursor.Color)

	// By default, the blink speed of the cursor is set to a default
	// internally.
	if m.styles.Cursor.Blink {
		if m.styles.Cursor.BlinkSpeed > 0 {
			m.virtualCursor.BlinkSpeed = m.styles.Cursor.BlinkSpeed
		}

		m.virtualCursor.SetMode(cursor.CursorBlink)

		return
	}

	m.virtualCursor.SetMode(cursor.CursorStatic)
}

func (m *Model) insertRunesFromUserInput(runes []rune) {
	runes, fits := m.fitCharLimit(m.san().Sanitize(runes))
	if !fits {
		return
	}

	lines := m.fitContentHeight(m.fitMaxLines(splitLines(runes)))
	if len(lines) == 0 {
		return
	}

	m.insertLines(lines)
}

func (m *Model) fitCharLimit(runes []rune) ([]rune, bool) {
	if m.CharLimit <= 0 {
		return runes, true
	}

	availSpace := m.CharLimit - m.Length()
	if availSpace <= 0 {
		return nil, false
	}

	return runes[:min(availSpace, len(runes))], true
}

func (m *Model) fitMaxLines(lines [][]rune) [][]rune {
	if m.MaxLines > 0 && len(m.value)+len(lines)-1 > m.MaxLines {
		allowedHeight := max(0, m.MaxLines-len(m.value)+1)
		lines = lines[:allowedHeight]
	}

	return lines
}

func (m *Model) fitContentHeight(lines [][]rune) [][]rune {
	if m.MaxContentHeight <= 0 {
		return lines
	}

	budget := m.MaxContentHeight - m.totalVisualLines()
	for len(lines) > 1 && m.visualLinesForInsert(lines) > budget {
		lines = lines[:len(lines)-1]
	}

	if m.visualLinesForInsert(lines) > budget {
		return nil
	}

	return lines
}

func (m *Model) insertLines(lines [][]rune) {
	tail := make([]rune, len(m.value[m.row][m.col:]))
	copy(tail, m.value[m.row][m.col:])

	m.value[m.row] = append(m.value[m.row][:m.col], lines[0]...)
	m.col += len(lines[0])

	m.insertExtraLines(lines[1:])

	m.value[m.row] = append(m.value[m.row], tail...)

	m.SetCursorColumn(m.col)
}

func (m *Model) insertExtraLines(extraLines [][]rune) {
	if len(extraLines) == 0 {
		return
	}

	m.value = m.gridWithRoomFor(len(extraLines))
	for _, l := range extraLines {
		m.row++
		m.value[m.row] = l
		m.col = len(l)
	}
}

func (m *Model) gridWithRoomFor(numExtraLines int) [][]rune {
	var newGrid [][]rune
	if cap(m.value) >= len(m.value)+numExtraLines {
		newGrid = m.value[:len(m.value)+numExtraLines]
	} else {
		newGrid = make([][]rune, len(m.value)+numExtraLines)
		copy(newGrid, m.value[:m.row+1])
	}

	copy(newGrid[m.row+1+numExtraLines:], m.value[m.row+1:])

	return newGrid
}

func (m *Model) setCursorLineRelative(delta int) {
	if delta == 0 {
		return
	}

	lineInfo := m.LineInfo()
	charOffset := max(m.lastCharOffset, lineInfo.CharOffset)
	m.lastCharOffset = charOffset

	if delta > 0 {
		m.moveDownLines(delta, lineInfo)
	} else {
		m.moveUpLines(-delta, lineInfo)
	}

	m.moveToCharOffset(charOffset)
	m.repositionView()
}

func (m *Model) moveDownLines(count int, lineInfo LineInfo) {
	for range count {
		if lineInfo.RowOffset+1 >= lineInfo.Height && m.row < len(m.value)-1 {
			m.row++
			m.col = 0
		} else {
			m.col = min(lineInfo.StartColumn+lineInfo.Width+trailingSpace, len(m.value[m.row])-1)
		}

		lineInfo = m.LineInfo()
	}
}

func (m *Model) moveUpLines(count int, lineInfo LineInfo) {
	for range count {
		if lineInfo.RowOffset <= 0 && m.row > 0 {
			m.row--
			m.col = len(m.value[m.row])
		} else {
			m.col = lineInfo.StartColumn - trailingSpace
		}

		lineInfo = m.LineInfo()
	}
}

func (m *Model) moveToCharOffset(charOffset int) {
	nli := m.LineInfo()
	m.col = nli.StartColumn

	if nli.Width <= 0 {
		return
	}

	offset := 0
	for offset < charOffset {
		if m.row >= len(m.value) || m.col >= len(m.value[m.row]) || offset >= nli.CharWidth-1 {
			break
		}

		char := m.value[m.row][m.col]

		next := offset + tabstop.RuneCellsAt(char, nli.StartCharOffset+offset)
		if next > charOffset && tabstop.IsTab(char) {
			break
		}

		offset = next
		m.col++
	}
}

func (m *Model) activeStyle() *StyleState {
	if m.focus {
		return &m.styles.Focused
	}

	return &m.styles.Blurred
}

func (m *Model) san() *runeutil.Sanitizer {
	if m.rsan == nil {
		m.rsan = runeutil.NewSanitizer()
	}

	return m.rsan
}

func (m *Model) deleteBeforeCursor() {
	m.value[m.row] = m.value[m.row][m.col:]
	m.SetCursorColumn(0)
}

func (m *Model) deleteAfterCursor() {
	m.value[m.row] = m.value[m.row][:m.col]
	m.SetCursorColumn(len(m.value[m.row]))
}

func (m *Model) transposeLeft() {
	if m.col == 0 || len(m.value[m.row]) < 2 {
		return
	}

	if m.col >= len(m.value[m.row]) {
		m.SetCursorColumn(m.col - 1)
	}

	m.value[m.row][m.col-1], m.value[m.row][m.col] = m.value[m.row][m.col], m.value[m.row][m.col-1]
	if m.col < len(m.value[m.row]) {
		m.SetCursorColumn(m.col + 1)
	}
}

func (m *Model) deleteWordLeft() {
	if m.col == 0 || len(m.value[m.row]) == 0 {
		return
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldCol := m.col

	m.SetCursorColumn(m.col - 1)

	for unicode.IsSpace(m.value[m.row][m.col]) {
		if m.col <= 0 {
			break
		}

		m.SetCursorColumn(m.col - 1)
	}

	for m.col > 0 {
		if unicode.IsSpace(m.value[m.row][m.col]) {
			// keep the previous space
			m.SetCursorColumn(m.col + 1)

			break
		}

		m.SetCursorColumn(m.col - 1)
	}

	if oldCol > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:m.col]
	} else {
		m.value[m.row] = append(m.value[m.row][:m.col], m.value[m.row][oldCol:]...)
	}
}

func (m *Model) deleteWordRight() {
	if m.col >= len(m.value[m.row]) || len(m.value[m.row]) == 0 {
		return
	}

	oldCol := m.col

	for m.col < len(m.value[m.row]) && unicode.IsSpace(m.value[m.row][m.col]) {
		m.SetCursorColumn(m.col + 1)
	}

	for m.col < len(m.value[m.row]) && !unicode.IsSpace(m.value[m.row][m.col]) {
		m.SetCursorColumn(m.col + 1)
	}

	if m.col > len(m.value[m.row]) {
		m.value[m.row] = m.value[m.row][:oldCol]
	} else {
		m.value[m.row] = append(m.value[m.row][:oldCol], m.value[m.row][m.col:]...)
	}

	m.SetCursorColumn(oldCol)
}

func (m *Model) characterRight() {
	if m.col < len(m.value[m.row]) {
		m.SetCursorColumn(m.col + 1)
	} else if m.row < len(m.value)-1 {
		m.row++
		m.CursorStart()
	}
}

func (m *Model) characterLeft() {
	if m.wrapToPreviousLineEnd() {
		return
	}

	m.stepLeft()
}

func (m *Model) characterLeftInsideLine() {
	m.wrapToPreviousLineEnd()
	m.stepLeft()
}

func (m *Model) wrapToPreviousLineEnd() bool {
	if m.col != 0 || m.row == 0 {
		return false
	}

	m.row--
	m.CursorEnd()

	return true
}

func (m *Model) stepLeft() {
	if m.col > 0 {
		m.SetCursorColumn(m.col - 1)
	}
}

func (m *Model) wordLeft() {
	for {
		oldRow, oldCol := m.row, m.col
		m.characterLeftInsideLine()

		if m.row == oldRow && m.col == oldCol {
			return
		}

		if m.col < len(m.value[m.row]) && !unicode.IsSpace(m.value[m.row][m.col]) {
			break
		}
	}

	for m.col > 0 {
		if unicode.IsSpace(m.value[m.row][m.col-1]) {
			break
		}

		m.SetCursorColumn(m.col - 1)
	}
}

func (m *Model) wordRight() {
	m.doWordRight(func(int, int) {})
}

func (m *Model) doWordRight(visit func(charIdx, pos int)) {
	for m.col >= len(m.value[m.row]) || unicode.IsSpace(m.value[m.row][m.col]) {
		if m.row == len(m.value)-1 && m.col == len(m.value[m.row]) {
			break
		}

		m.characterRight()
	}

	charIdx := 0

	for m.col < len(m.value[m.row]) {
		if unicode.IsSpace(m.value[m.row][m.col]) {
			break
		}

		visit(charIdx, m.col)
		m.SetCursorColumn(m.col + 1)

		charIdx++
	}
}

func (m *Model) uppercaseRight() {
	m.doWordRight(func(_, i int) {
		m.value[m.row][i] = unicode.ToUpper(m.value[m.row][i])
	})
}

func (m *Model) lowercaseRight() {
	m.doWordRight(func(_, i int) {
		m.value[m.row][i] = unicode.ToLower(m.value[m.row][i])
	})
}

func (m *Model) capitalizeRight() {
	m.doWordRight(func(charIdx, i int) {
		if charIdx == 0 {
			m.value[m.row][i] = unicode.ToTitle(m.value[m.row][i])
		}
	})
}

func (m *Model) repositionView() {
	minimum := m.viewport.YOffset()

	maximum := minimum + m.viewport.Height() - 1
	if row := m.cursorLineNumber(); row < minimum {
		m.viewport.ScrollUp(minimum - row)
	} else if row > maximum {
		m.viewport.ScrollDown(row - maximum)
	}
}

func (m *Model) prepareForMsg() {
	if m.value[m.row] == nil {
		m.value[m.row] = make([]rune, 0)
	}

	if m.MaxHeight > 0 && m.MaxHeight != m.cache.Capacity() {
		m.cache = memoization.NewMemoCache[line, []wrappedLine](m.MaxHeight)
	}
}

func (m *Model) handleMsg(msg tea.Msg) keyOutcome {
	switch msg := msg.(type) {
	case tea.PasteMsg:
		m.deleteSelection()
		m.insertRunesFromUserInput([]rune(msg.Content))
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return proceed()
}

func (m *Model) handleKey(msg tea.KeyPressMsg) keyOutcome {
	for _, action := range m.keyActions {
		if key.Matches(msg, action.binding(&m.KeyMap)) {
			return action.handle(m)
		}
	}

	if msg.Text != "" {
		m.deleteSelection()
		m.insertRunesFromUserInput([]rune(msg.Text))
	}

	return proceed()
}

func (m *Model) deleteToLineEnd() {
	m.col = clamp(m.col, 0, len(m.value[m.row]))
	if m.col >= len(m.value[m.row]) {
		m.mergeLineBelow(m.row)

		return
	}

	m.deleteAfterCursor()
}

func (m *Model) deleteToLineStart() {
	m.col = clamp(m.col, 0, len(m.value[m.row]))
	if m.col <= 0 {
		m.mergeLineAbove(m.row)

		return
	}

	m.deleteBeforeCursor()
}

func (m *Model) deleteCharacterBackward() {
	m.col = clamp(m.col, 0, len(m.value[m.row]))
	if m.col <= 0 {
		m.mergeLineAbove(m.row)

		return
	}

	if len(m.value[m.row]) > 0 {
		m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
		m.stepLeft()
	}
}

func (m *Model) deleteCharacterForward() {
	if len(m.value[m.row]) > 0 && m.col < len(m.value[m.row]) {
		m.value[m.row] = slices.Delete(m.value[m.row], m.col, m.col+1)
	}

	if m.col >= len(m.value[m.row]) {
		m.mergeLineBelow(m.row)
	}
}

func (m *Model) deleteWordBackward() {
	if m.col <= 0 {
		m.mergeLineAbove(m.row)

		return
	}

	m.deleteWordLeft()
}

func (m *Model) deleteWordForward() {
	m.col = clamp(m.col, 0, len(m.value[m.row]))
	if m.col >= len(m.value[m.row]) {
		m.mergeLineBelow(m.row)

		return
	}

	m.deleteWordRight()
}

func (m *Model) insertNewline() keyOutcome {
	m.deleteSelection()

	if m.atContentLimit() {
		return finishUpdate()
	}

	m.col = clamp(m.col, 0, len(m.value[m.row]))
	m.splitLine(m.row, m.col)

	return proceed()
}

func (m *Model) refresh(msg tea.Msg, oldRow, oldCol int) tea.Cmd {
	m.recalculateHeight()

	// Make sure we set the content of the viewport before updating it.
	view := m.contentView()
	m.viewport.SetContent(view)
	scrolled, viewportCmd := m.viewport.Update(msg)
	m.viewport = &scrolled

	cmds := []tea.Cmd{viewportCmd}
	if m.useVirtualCursor {
		cmds = append(cmds, m.updateVirtualCursor(msg, oldRow, oldCol))
	}

	m.repositionView()

	return tea.Batch(cmds...)
}

func (m *Model) updateVirtualCursor(msg tea.Msg, oldRow, oldCol int) tea.Cmd {
	var cmd tea.Cmd

	m.virtualCursor, cmd = m.virtualCursor.Update(msg)

	// If the cursor has moved, reset the blink state. This is a small UX
	// nuance that makes cursor movement obvious and feel snappy.
	newRow, newCol := m.cursorLineNumber(), m.col
	if (newRow != oldRow || newCol != oldCol) && m.virtualCursor.Mode() == cursor.CursorBlink {
		m.virtualCursor.IsBlinked = false
		cmd = m.virtualCursor.Blink()
	}

	return cmd
}

func (m *Model) contentView() string {
	if m.Value() == "" && m.row == 0 && m.col == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	m.virtualCursor.TextStyle = m.activeStyle().computedCursorLine()

	view := contentRenderer{
		model:       m,
		styles:      m.activeStyle(),
		lineInfo:    m.LineInfo(),
		out:         strings.Builder{},
		displayLine: 0,
	}
	for row, line := range m.value {
		view.writeLine(row, line)
	}

	view.writeEndOfBuffer()

	return view.out.String()
}

func (m *Model) promptView(displayLine int) string {
	prompt := m.Prompt
	if m.promptFunc != nil {
		prompt = m.promptFunc(PromptInfo{
			LineNumber: displayLine,
			Focused:    m.focus,
		})

		width := lipgloss.Width(prompt)
		if width < m.promptWidth {
			prompt = fmt.Sprintf("%*s%s", m.promptWidth-width, "", prompt)
		}
	}

	return m.activeStyle().computedPrompt().Render(prompt)
}

func (m *Model) lineNumberFor(row, number int) string {
	if m.row == row {
		return m.cursorLineNumberView(number)
	}

	return m.lineNumberView(number)
}

func (m *Model) lineNumberView(number int) string {
	styles := m.activeStyle()

	return m.renderLineNumber(number, styles.computedText(), styles.computedLineNumber())
}

func (m *Model) cursorLineNumberView(number int) string {
	styles := m.activeStyle()

	return m.renderLineNumber(number, styles.computedCursorLine(), styles.computedCursorLineNumber())
}

// If the argument is less than 0, a space styled as a line number is returned
// instead. Such cases are used for soft-wrapped lines.
func (m *Model) renderLineNumber(number int, textStyle, lineNumberStyle lipgloss.Style) string {
	if !m.ShowLineNumbers {
		return ""
	}

	str := " "
	if number > 0 {
		str = strconv.Itoa(number)
	}

	digits := len(strconv.Itoa(m.MaxHeight))
	str = fmt.Sprintf(" %*v ", digits, str)

	return textStyle.Render(lineNumberStyle.Render(str))
}

func (m *Model) placeholderView() string {
	pwordwrap := ansi.Wordwrap(m.Placeholder, m.width, "")
	// hard wrap lines (handles lines that could not be word wrapped)
	pwrap := ansi.Hardwrap(pwordwrap, m.width, true)
	plines := strings.Split(strings.TrimSpace(pwrap), "\n")

	var out strings.Builder
	for displayLine := range m.height {
		out.WriteString(m.placeholderLine(plines, displayLine))
		out.WriteRune('\n')
	}

	m.viewport.SetContent(out.String())

	return m.activeStyle().Base.Render(m.viewport.View())
}

func (m *Model) placeholderLine(plines []string, displayLine int) string {
	lineStyle := m.activeStyle().computedPlaceholder()
	if len(plines) > displayLine {
		lineStyle = m.activeStyle().computedCursorLine()
	}

	prompt := lineStyle.Render(m.promptView(displayLine))

	return prompt + m.placeholderLineNumber(plines, displayLine) + m.placeholderText(plines, displayLine, lineStyle)
}

func (m *Model) placeholderLineNumber(plines []string, displayLine int) string {
	// when show line numbers enabled:
	// - render line number for only the cursor line
	// - indent other placeholder lines
	// this is consistent with vim with line numbers enabled
	switch {
	case displayLine == 0:
		return m.cursorLineNumberView(1)
	case len(plines) > displayLine:
		return m.cursorLineNumberView(0)
	default:
		return ""
	}
}

func (m *Model) placeholderText(plines []string, displayLine int, lineStyle lipgloss.Style) string {
	styles := m.activeStyle()

	switch {
	case displayLine == 0:
		return m.firstPlaceholderLine(plines[0], lineStyle)
	case len(plines) > displayLine:
		placeholderLine := plines[displayLine]
		gap := strings.Repeat(" ", max(0, m.width-uniseg.StringWidth(placeholderLine)))

		return lineStyle.Render(placeholderLine + gap)
	default:
		return styles.computedEndOfBuffer().Render(string(m.EndOfBufferCharacter))
	}
}

func (m *Model) firstPlaceholderLine(placeholder string, lineStyle lipgloss.Style) string {
	styles := m.activeStyle()

	m.virtualCursor.TextStyle = styles.computedPlaceholder()

	ch, rest, _, _ := uniseg.FirstGraphemeClusterInString(placeholder, 0)
	m.virtualCursor.SetChar(ch)

	// extend the first line with spaces to fill the width, so that
	// the entire line is filled when cursorline is enabled.
	gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(placeholder)))

	return lineStyle.Render(m.virtualCursor.View()) +
		lineStyle.Render(styles.computedPlaceholder().Render(rest)) +
		lineStyle.Render(gap)
}

func (m *Model) deleteSelection() {
	selected, ok := m.Selection()
	if !ok {
		return
	}

	start, end := selected.Start, selected.End
	if start.Row == end.Row {
		m.deleteWithinLine(start, end)
	} else {
		m.deleteAcrossLines(start, end)
	}

	m.ClearSelection()
}

func (m *Model) deleteWithinLine(start, end Position) {
	line := m.value[start.Row]
	endCol := min(end.Col, len(line))
	startCol := min(start.Col, endCol)
	m.value[start.Row] = append(line[:startCol], line[endCol:]...)
	m.row = start.Row
	m.SetCursorColumn(startCol)
}

func (m *Model) deleteAcrossLines(start, end Position) {
	headLine := m.value[start.Row]
	tailLine := m.value[end.Row]
	endCol := min(end.Col, len(tailLine))
	startCol := min(start.Col, len(headLine))

	merged := make([]rune, 0, startCol+len(tailLine)-endCol)
	merged = append(merged, headLine[:startCol]...)
	merged = append(merged, tailLine[endCol:]...)

	newValue := make([][]rune, 0, len(m.value)-(end.Row-start.Row))
	newValue = append(newValue, m.value[:start.Row]...)
	newValue = append(newValue, merged)
	newValue = append(newValue, m.value[end.Row+1:]...)
	m.value = newValue

	m.row = start.Row
	m.SetCursorColumn(startCol)
}

func (m *Model) startKeyboardSelection() {
	if !m.hasSelection {
		m.selAnchor = Position{Row: m.row, Col: m.col}
	}

	m.hasSelection = true
	m.selHead = Position{Row: m.row, Col: m.col}
}

func (m *Model) updateKeyboardSelection() {
	m.selHead = Position{Row: m.row, Col: m.col}
}

func (m *Model) memoizedWrap(runes []rune, width int) []wrappedLine {
	input := line{runes: runes, width: width}
	if v, ok := m.cache.Get(input); ok {
		return v
	}

	v := wrap(runes, width)
	m.cache.Set(input, v)

	return v
}

// This accounts for soft wrapped lines.
func (m *Model) cursorLineNumber() int {
	line := 0
	for i := range m.row {
		line += len(m.memoizedWrap(m.value[i], m.width))
	}

	line += m.LineInfo().RowOffset

	return line
}

func (m *Model) totalVisualLines() int {
	n := 0
	for _, line := range m.value {
		n += len(m.memoizedWrap(line, m.width))
	}

	return n
}

func (m *Model) recalculateHeight() {
	if !m.DynamicHeight {
		return
	}

	minH := max(m.MinHeight, minHeight)
	total := m.totalVisualLines()

	height := max(total, minH)
	if m.MaxHeight > 0 {
		height = min(height, m.MaxHeight)
	}

	if maxOffset := total - height; m.viewport.YOffset() > maxOffset {
		m.viewport.SetYOffset(max(0, maxOffset))
	}

	m.SetHeight(height)
}

func (m *Model) atContentLimit() bool {
	if m.MaxContentHeight > 0 {
		return m.totalVisualLines() >= m.MaxContentHeight
	}

	return m.MaxHeight > 0 && len(m.value) >= m.MaxHeight
}

// visualLinesForInsert estimates how many additional visual lines would result
// from inserting the given lines at the current cursor position. The first
// element merges into the current line; subsequent elements become new lines.
func (m *Model) visualLinesForInsert(lines [][]rune) int {
	if len(lines) == 0 {
		return 0
	}

	currentRowVisual := len(m.memoizedWrap(m.value[m.row], m.width))

	merged := make([]rune, 0, m.col+len(lines[0]))
	merged = append(merged, m.value[m.row][:m.col]...)
	merged = append(merged, lines[0]...)

	if len(lines) == 1 {
		merged = append(merged, m.value[m.row][m.col:]...)
	}

	delta := len(m.memoizedWrap(merged, m.width)) - currentRowVisual

	for i, content := range lines {
		if i == len(lines)-1 {
			content = append(content, m.value[m.row][m.col:]...)
		}

		delta += len(m.memoizedWrap(content, m.width))
	}

	return delta
}

func (m *Model) mergeLineBelow(row int) {
	if row >= len(m.value)-1 {
		return
	}

	m.value[row] = append(m.value[row], m.value[row+1]...)

	for i := row + 1; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	m.col = len(m.value[row-1])
	m.row--

	m.value[row-1] = append(m.value[row-1], m.value[row]...)

	for i := row; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m *Model) splitLine(row, col int) {
	head, tailSrc := m.value[row][:col], m.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	m.value = append(m.value[:row+1], m.value[row:]...)

	m.value[row] = head
	m.value[row+1] = tail

	m.col = 0
	m.row++
}

type keyOutcome struct {
	finished bool
}

type bindingOf func(*KeyMap) key.Binding

type keyAction struct {
	binding bindingOf
	handle  func(*Model) keyOutcome
}

func newKeyActions() []keyAction {
	return []keyAction{
		deleting(func(k *KeyMap) key.Binding { return k.DeleteAfterCursor }, (*Model).deleteToLineEnd),
		deleting(func(k *KeyMap) key.Binding { return k.DeleteBeforeCursor }, (*Model).deleteToLineStart),
		deleting(func(k *KeyMap) key.Binding { return k.DeleteCharacterBackward }, (*Model).deleteCharacterBackward),
		deleting(func(k *KeyMap) key.Binding { return k.DeleteCharacterForward }, (*Model).deleteCharacterForward),
		deleting(func(k *KeyMap) key.Binding { return k.DeleteWordBackward }, (*Model).deleteWordBackward),
		deleting(func(k *KeyMap) key.Binding { return k.DeleteWordForward }, (*Model).deleteWordForward),
		{binding: func(k *KeyMap) key.Binding { return k.InsertNewline }, handle: (*Model).insertNewline},
		moving(func(k *KeyMap) key.Binding { return k.LineEnd }, (*Model).CursorEnd),
		moving(func(k *KeyMap) key.Binding { return k.LineStart }, (*Model).CursorStart),
		moving(func(k *KeyMap) key.Binding { return k.CharacterForward }, (*Model).characterRight),
		moving(func(k *KeyMap) key.Binding { return k.LineNext }, (*Model).CursorDown),
		moving(func(k *KeyMap) key.Binding { return k.WordForward }, (*Model).wordRight),
		moving(func(k *KeyMap) key.Binding { return k.CharacterBackward }, (*Model).characterLeft),
		moving(func(k *KeyMap) key.Binding { return k.LinePrevious }, (*Model).CursorUp),
		moving(func(k *KeyMap) key.Binding { return k.WordBackward }, (*Model).wordLeft),
		moving(func(k *KeyMap) key.Binding { return k.InputBegin }, (*Model).MoveToBegin),
		moving(func(k *KeyMap) key.Binding { return k.InputEnd }, (*Model).MoveToEnd),
		moving(func(k *KeyMap) key.Binding { return k.PageUp }, (*Model).PageUp),
		moving(func(k *KeyMap) key.Binding { return k.PageDown }, (*Model).PageDown),
		editing(func(k *KeyMap) key.Binding { return k.LowercaseWordForward }, (*Model).lowercaseRight),
		editing(func(k *KeyMap) key.Binding { return k.UppercaseWordForward }, (*Model).uppercaseRight),
		editing(func(k *KeyMap) key.Binding { return k.CapitalizeWordForward }, (*Model).capitalizeRight),
		editing(func(k *KeyMap) key.Binding { return k.TransposeCharacterBackward }, (*Model).transposeLeft),
		selecting(func(k *KeyMap) key.Binding { return k.SelectCharacterForward }, (*Model).characterRight),
		selecting(func(k *KeyMap) key.Binding { return k.SelectCharacterBackward }, (*Model).characterLeft),
		selecting(func(k *KeyMap) key.Binding { return k.SelectWordForward }, (*Model).wordRight),
		selecting(func(k *KeyMap) key.Binding { return k.SelectWordBackward }, (*Model).wordLeft),
		selecting(func(k *KeyMap) key.Binding { return k.SelectLineUp }, (*Model).CursorUp),
		selecting(func(k *KeyMap) key.Binding { return k.SelectLineDown }, (*Model).CursorDown),
		editing(func(k *KeyMap) key.Binding { return k.SelectAll }, (*Model).SelectAll),
	}
}

type contentRenderer struct {
	model       *Model
	styles      *StyleState
	lineInfo    LineInfo
	out         strings.Builder
	displayLine int
}

type wrappedSegment struct {
	row        int
	index      int
	base       int
	lineLength int
	wrapped    wrappedLine
}

func (v *contentRenderer) writeLine(row int, line []rune) {
	style := v.styles.computedText()
	if v.model.row == row {
		style = v.styles.computedCursorLine()
	}

	// wrappedBase is the index, within the logical line, of the first rune
	// of the current wrapped segment. It is what maps a segment back onto
	// selection coordinates.
	wrappedBase := 0

	for index, wrapped := range v.model.memoizedWrap(line, v.model.width) {
		segment := wrappedSegment{row: row, index: index, base: wrappedBase, lineLength: len(line), wrapped: wrapped}
		wrappedBase += v.writeSegment(segment, style)
	}
}

func (v *contentRenderer) writeSegment(segment wrappedSegment, style lipgloss.Style) int {
	model := v.model

	v.out.WriteString(style.Render(model.promptView(v.displayLine)))
	v.displayLine++

	lineNumber := -1
	if segment.index == 0 {
		lineNumber = segment.row + 1
	}

	v.out.WriteString(model.lineNumberFor(segment.row, lineNumber))

	wrappedLength := len(segment.wrapped.runes)
	// If the trailing space or tab causes the line to be wider than the
	// width, we should not draw it to the screen since it will result
	// in an extra space at the end of the line which can look off when
	// the cursor line is showing.
	if segment.wrapped.width() > model.width {
		segment.wrapped = segment.wrapped.withoutTrailingWhitespace()
	}

	v.writeSegmentText(segment, style)

	v.out.WriteString(style.Render(strings.Repeat(" ", max(0, model.width-segment.wrapped.width()))))
	v.out.WriteRune('\n')

	return wrappedLength
}

func (v *contentRenderer) writeSegmentText(segment wrappedSegment, style lipgloss.Style) {
	// A selection covering this segment takes precedence over the
	// inline cursor: during a drag the highlight is the meaningful
	// affordance, and mixing the two would double-style the same cell.
	if span, selected := v.model.selectionSpanFor(segment.row, segment.base, len(segment.wrapped.runes)); selected {
		drawn := segment.wrapped.split(span.from, span.to)
		v.out.WriteString(style.Render(drawn.before))
		v.out.WriteString(v.styles.computedSelection().Render(drawn.inside))
		v.out.WriteString(style.Render(drawn.after))

		return
	}

	if v.model.row == segment.row && v.lineInfo.RowOffset == segment.index {
		v.writeCursorSegment(segment, style)

		return
	}

	v.out.WriteString(style.Render(segment.wrapped.drawn()))
}

func (v *contentRenderer) writeCursorSegment(segment wrappedSegment, style lipgloss.Style) {
	model := v.model
	cursorAt := v.lineInfo.ColumnOffset
	drawn := segment.wrapped.split(cursorAt, min(cursorAt+1, len(segment.wrapped.runes)))

	v.out.WriteString(style.Render(drawn.before))

	if model.col >= segment.lineLength && v.lineInfo.CharOffset >= model.width {
		model.virtualCursor.SetChar(" ")
		v.out.WriteString(model.virtualCursor.View())

		return
	}

	model.virtualCursor.SetChar(drawn.inside)
	v.out.WriteString(style.Render(model.virtualCursor.View()))
	v.out.WriteString(style.Render(drawn.after))
}

func (v *contentRenderer) writeEndOfBuffer() {
	model := v.model

	// Always show at least `m.Height` lines at all times.
	for range model.height {
		v.out.WriteString(model.promptView(v.displayLine))
		v.displayLine++

		leftGutter := string(model.EndOfBufferCharacter)
		rightGapWidth := model.Width() - uniseg.StringWidth(leftGutter)
		rightGap := strings.Repeat(" ", max(0, rightGapWidth))
		v.out.WriteString(v.styles.computedEndOfBuffer().Render(leftGutter + rightGap))
		v.out.WriteRune('\n')
	}
}

func proceed() keyOutcome {
	return keyOutcome{finished: false}
}

func finishUpdate() keyOutcome {
	return keyOutcome{finished: true}
}

func editing(binding bindingOf, change func(*Model)) keyAction {
	return keyAction{
		binding: binding,
		handle: func(model *Model) keyOutcome {
			change(model)

			return proceed()
		},
	}
}

func deleting(binding bindingOf, remove func(*Model)) keyAction {
	return editing(binding, func(model *Model) {
		if model.HasSelection() {
			model.deleteSelection()

			return
		}

		remove(model)
	})
}

func moving(binding bindingOf, move func(*Model)) keyAction {
	return editing(binding, func(model *Model) {
		model.ClearSelection()
		move(model)
	})
}

func selecting(binding bindingOf, move func(*Model)) keyAction {
	return editing(binding, func(model *Model) {
		model.startKeyboardSelection()
		move(model)
		model.updateKeyboardSelection()
	})
}

func splitLines(runes []rune) [][]rune {
	var lines [][]rune

	lstart := 0

	for index := range runes {
		if runes[index] == '\n' {
			// Beware to clamp the max capacity of the slice, to ensure no
			// data from different rows get overwritten when later edits
			// will modify this line.
			lines = append(lines, runes[lstart:index:index])
			lstart = index + 1
		}
	}

	if lstart <= len(runes) {
		lines = append(lines, runes[lstart:])
	}

	return lines
}

func wrap(runes []rune, width int) []wrappedLine {
	wrapper := wrapping{width: width, lines: []wrappedLine{{runes: []rune{}, startCell: 0}}, word: []rune{}, row: 0}

	for _, r := range runes {
		wrapper.add(r)
	}

	return wrapper.finish()
}

type wrapping struct {
	width int
	lines []wrappedLine
	word  []rune
	row   int
}

func (w *wrapping) add(char rune) {
	if unicode.IsSpace(char) {
		w.placeWordAnd(char)

		return
	}

	w.word = append(w.word, char)
	w.breakOverlongWord()
}

func (w *wrapping) placeWordAnd(space rune) {
	if len(w.lines[w.row].runes) > 0 && w.lines[w.row].widthWith(w.word, []rune{space}) > w.width {
		w.startLine()
	}

	w.placeWord()
	w.place(space)
}

func (w *wrapping) breakOverlongWord() {
	// If the last character is a double-width rune, then we may not be able to add it to this line
	// as it might cause us to go past the width.
	lastCharLen := rw.RuneWidth(w.word[len(w.word)-1])
	if uniseg.StringWidth(string(w.word))+lastCharLen <= w.width {
		return
	}

	if len(w.lines[w.row].runes) > 0 {
		w.startLine()
	}

	w.placeWord()
}

func (w *wrapping) finish() []wrappedLine {
	if w.lines[w.row].widthWith(w.word) >= w.width {
		w.startLine()
	}

	w.placeWord()
	// We add an extra space at the end of the line to account for the
	// trailing space at the end of the previous soft-wrapped lines so that
	// behaviour when navigating is consistent and so that we don't need to
	// continually add edges to handle the last line of the wrapped input.
	w.place(' ')

	return w.lines
}

func (w *wrapping) placeWord() {
	w.place(w.word...)
	w.word = nil
}

func (w *wrapping) place(runes ...rune) {
	w.lines[w.row].runes = append(w.lines[w.row].runes, runes...)
}

func (w *wrapping) startLine() {
	w.lines = append(w.lines, w.lines[w.row].following())
	w.row++
}

func numDigits(number int) int {
	if number == 0 {
		return 1
	}

	count := 0

	num := abs(number)
	for num > 0 {
		count++
		num /= 10
	}

	return count
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}

	return min(high, max(low, v))
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
