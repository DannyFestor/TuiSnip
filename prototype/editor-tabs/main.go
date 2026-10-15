// PROTOTYPE for wayfinder ticket "An editor that keeps tabs and highlights while editing"
// (DannyFestor/TuiSnip#206). Throwaway: a forked Bubbles v2.2.1 textarea that keeps tabs,
// with the bubbles#1030 highlighter patch and chroma. Not TuiSnip code.
//
//	go run .                     the built-in sample (Go with tabs, a Makefile rule, TSV)
//	go run . -file path          any file, Language from its name
//	go run . -kib 256            a generated Go Fragment at the 256 KiB Fragment cap
//	go run . -lines 60000        60,000 short lines, far past the old 10,000-line cap
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tuisnip-editor-prototype/textarea"
)

//go:embed sample.txt
var sample string

var (
	tabWidths    = []int{4, 8, 2}
	cursorModes  = []string{"whole", "first", "last"}
	tabKeyModes  = []string{"insert \\t", "indent line with \\t", "indent line with 4 spaces"}
	glyphs       = []string{"", "→", "»"}
	lineCaps     = []int{0, 10000}
	bannerLook   = lipgloss.NewStyle().Background(lipgloss.Color("#b58900")).Foreground(lipgloss.Color("#000000")).Bold(true)
	stateLook    = lipgloss.NewStyle().Foreground(lipgloss.Color("#93a1a1"))
	knobLook     = lipgloss.NewStyle().Foreground(lipgloss.Color("#268bd2"))
	feedbackLook = lipgloss.NewStyle().Foreground(lipgloss.Color("#cb4b16"))
)

type model struct {
	area      textarea.Model
	highlight *highlighter
	source    string

	tabWidth, cursorMode, tabKeyMode, glyph, lineCap, highlightMode int

	lastValue  string
	updateTime time.Duration
	viewTime   time.Duration
	valueTime  time.Duration
	lexTime    time.Duration
	feedback   string
	pendingLex string
	lexSeq     int
}

func main() {
	file := flag.String("file", "", "file to open")
	kib := flag.Int("kib", 0, "generate a Go Fragment of this many KiB")
	lines := flag.Int("lines", 0, "generate this many short lines")
	flag.Parse()

	content, source, language := load(*file, *kib, *lines)
	m := newModel(content, source, language)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func load(file string, kib, lines int) (content, source, language string) {
	switch {
	case file != "":
		raw, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return string(raw), file, languageOf(file)
	case kib > 0:
		return generatedGo(kib * 1024), fmt.Sprintf("generated %d KiB of Go", kib), "go"
	case lines > 0:
		return shortLines(lines), fmt.Sprintf("generated %d short lines", lines), "go"
	}
	return sample, "built-in sample", "go"
}

func languageOf(file string) string {
	if strings.EqualFold(filepath.Base(file), "Makefile") {
		return "makefile"
	}
	return strings.TrimPrefix(filepath.Ext(file), ".")
}

func newModel(content, source, language string) model {
	area := textarea.New()
	area.Prompt = ""
	area.MaxHeight = 0
	area.ShowLineNumbers = true
	m := model{area: area, highlight: newHighlighter(language), source: source}
	m.applyKnobs()
	m.area.SetValue(content)
	m.area.MoveToBegin()
	m.area.Focus()
	m.area.SetHighlighter(m.highlight)
	m.relexIfChanged()
	m.lexTime = m.highlight.relex(m.lastValue)
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.area.SetWidth(msg.Width)
		m.area.SetHeight(max(1, msg.Height-5))
		return m, nil
	case tea.KeyPressMsg:
		if handled, quit := m.handleKnob(msg.String()); handled {
			if quit {
				return m, tea.Quit
			}
			return m, m.afterChange()
		}
	case lexDueMsg:
		return m, m.lexInBackground(msg)
	case lexedMsg:
		m.applyLex(msg)
		return m, nil
	}

	start := time.Now()
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	m.updateTime = time.Since(start)
	return m, tea.Batch(cmd, m.afterChange())
}

const lexPause = 150 * time.Millisecond

type lexDueMsg struct{ seq int }

type lexedMsg struct {
	document string
	ranges   [][]lipgloss.Range
	took     time.Duration
}

// afterChange re-lexes synchronously in "document" mode, or schedules a lex for when typing
// pauses in "background" mode; until it lands, the previous colours stay.
func (m *model) afterChange() tea.Cmd {
	m.relexIfChanged()
	if m.highlight.mode != "background" || m.lastValue == m.highlight.lexed || m.lastValue == m.pendingLex {
		return nil
	}
	m.pendingLex = m.lastValue
	m.lexSeq++
	seq := m.lexSeq
	return tea.Tick(lexPause, func(time.Time) tea.Msg { return lexDueMsg{seq: seq} })
}

func (m *model) lexInBackground(due lexDueMsg) tea.Cmd {
	if due.seq != m.lexSeq {
		return nil
	}
	document, h := m.lastValue, m.highlight
	return func() tea.Msg {
		start := time.Now()
		ranges := h.lex(document)
		return lexedMsg{document: document, ranges: ranges, took: time.Since(start)}
	}
}

func (m *model) applyLex(lexed lexedMsg) {
	if lexed.document != m.lastValue {
		return
	}
	m.highlight.cached = lexed.ranges
	m.highlight.lexed = lexed.document
	m.lexTime = lexed.took
}

func (m *model) handleKnob(key string) (handled, quit bool) {
	m.feedback = ""
	switch key {
	case "ctrl+c", "esc":
		return true, true
	case "tab":
		m.pressTab()
	case "shift+tab":
		m.area.DedentLine()
	case "f2":
		m.tabWidth = next(m.tabWidth, len(tabWidths))
	case "f3":
		textarea.TabStops = !textarea.TabStops
	case "f4":
		m.cursorMode = next(m.cursorMode, len(cursorModes))
	case "f5":
		textarea.PerCellTabs = !textarea.PerCellTabs
	case "f6":
		m.tabKeyMode = next(m.tabKeyMode, len(tabKeyModes))
	case "f7":
		m.highlightMode = next(m.highlightMode, len(highlightModes))
	case "f8":
		m.glyph = next(m.glyph, len(glyphs))
	case "f9":
		m.capture()
	case "f10":
		m.lineCap = next(m.lineCap, len(lineCaps))
	case "ctrl+r":
		textarea.WindowedRender = !textarea.WindowedRender
	default:
		return false, false
	}
	m.applyKnobs()
	return true, false
}

func next(i, n int) int {
	return (i + 1) % n
}

func (m *model) applyKnobs() {
	textarea.TabWidth = tabWidths[m.tabWidth]
	textarea.CursorOnTab = cursorModes[m.cursorMode]
	textarea.TabGlyph = glyphs[m.glyph]
	m.highlight.mode = highlightModes[m.highlightMode]
	m.area.MaxLines = lineCaps[m.lineCap]
	m.area.ResetWrapCache()
}

func (m *model) pressTab() {
	switch m.tabKeyMode {
	case 0:
		m.area.InsertString("\t")
	case 1:
		m.indentLine("\t")
	case 2:
		m.indentLine("    ")
	}
}

func (m *model) indentLine(indent string) {
	column := m.area.Column()
	m.area.CursorStart()
	m.area.InsertString(indent)
	m.area.SetCursorColumn(column + len([]rune(indent)))
}

// capture does what TuiSnip's Capture does: read the clipboard and SetValue it into the area.
func (m *model) capture() {
	raw, err := exec.Command("pbpaste").Output()
	if err != nil {
		m.feedback = "pbpaste failed: " + err.Error()
		return
	}
	content := string(raw)
	m.area.SetValue(content)
	m.area.MoveToBegin()
	m.source = fmt.Sprintf("Capture: %d bytes, %d tabs from the clipboard", len(content), strings.Count(content, "\t"))
	if got := m.area.Value(); got != content {
		m.feedback = fmt.Sprintf("Capture changed the content: %d bytes in, %d bytes kept", len(content), len(got))
	}
}

func (m *model) relexIfChanged() {
	start := time.Now()
	value := m.area.Value()
	m.valueTime = time.Since(start)
	m.lastValue = value
	if m.highlight.mode == "document" && value != m.highlight.lexed {
		m.lexTime = m.highlight.relex(value)
	}
}

func (m model) View() tea.View {
	m.highlight.lineLexTime = 0
	start := time.Now()
	area := m.area.View()
	m.viewTime = time.Since(start)

	body := lipgloss.JoinVertical(lipgloss.Left,
		bannerLook.Render(" PROTOTYPE · editor that keeps tabs · "+m.source+" · "+m.highlight.language()+" "),
		area,
		stateLook.Render(m.stateLine()),
		knobLook.Render(m.knobLine()),
	)
	view := tea.NewView(body)
	view.AltScreen = true
	return view
}

func (m model) stateLine() string {
	info := m.area.LineInfo()
	under := "end of line"
	switch r := m.area.RuneUnderCursor(); r {
	case -1:
	case '\t':
		under = fmt.Sprintf("\\t (cell %d)", m.area.TabCell())
	default:
		under = fmt.Sprintf("%q", r)
	}
	lineCap := "none"
	if c := lineCaps[m.lineCap]; c > 0 {
		lineCap = fmt.Sprint(c)
	}
	state := fmt.Sprintf("line %d/%d · rune col %d · cell col %d · on %s · %d tabs · line cap %s · update %s · view %s · Value() %s · lex %s",
		m.area.Line()+1, m.area.LineCount(), m.area.Column(), info.CharOffset, under,
		strings.Count(m.lastValue, "\t"), lineCap,
		ms(m.updateTime), ms(m.viewTime), ms(m.valueTime), ms(m.lexTime))
	if m.feedback != "" {
		state += "  " + feedbackLook.Render(m.feedback)
	}
	return state
}

func (m model) knobLine() string {
	stops := "fixed"
	if textarea.TabStops {
		stops = "stops"
	}
	move := "per tab"
	if textarea.PerCellTabs {
		move = "per cell"
	}
	glyph := "spaces"
	if g := glyphs[m.glyph]; g != "" {
		glyph = g
	}
	render := "every line"
	if textarea.WindowedRender {
		render = "near the viewport"
	}
	return fmt.Sprintf("F2 width %d · F3 %s · F4 cursor %s · F5 move %s · F6 tab key: %s\nF7 highlight %s · F8 tab shows %s · F9 Capture from clipboard · F10 line cap · ctrl+r render %s · shift+tab dedent · esc quit",
		tabWidths[m.tabWidth], stops, cursorModes[m.cursorMode], move, tabKeyModes[m.tabKeyMode],
		highlightModes[m.highlightMode], glyph, render)
}

func ms(d time.Duration) string {
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
}

func generatedGo(bytes int) string {
	var b strings.Builder
	b.WriteString("package generated\n\n")
	for i := 0; b.Len() < bytes; i++ {
		fmt.Fprintf(&b, `/*
	Handler%[1]d serves block %[1]d.
	This comment spans lines, which a per-line lexer can't know.
*/
func Handler%[1]d(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {	// not a GET
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	for i := 0; i < %[1]d; i++ {
		fmt.Fprintf(w, "%%d\t%%s\n", i, `+"`raw\n\tstring`"+`)
	}
}

`, i)
	}
	return b.String()
}

func shortLines(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "\tx%d\n", i%10)
	}
	return b.String()
}
