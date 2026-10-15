package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	stock "charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"tuisnip-editor-prototype/textarea"
)

func press(m model, key string) model {
	var msg tea.KeyPressMsg
	switch key {
	case "right":
		msg = tea.KeyPressMsg{Code: tea.KeyRight}
	case "left":
		msg = tea.KeyPressMsg{Code: tea.KeyLeft}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "tab":
		msg = tea.KeyPressMsg{Code: tea.KeyTab}
	default:
		msg = tea.KeyPressMsg{Code: []rune(key)[0], Text: key}
	}
	next, _ := m.Update(msg)
	return next.(model)
}

func sized(m model, w, h int) model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(model)
}

func screen(m model) string {
	return ansi.Strip(m.View().Content)
}

func TestSampleKeepsTabsAndRenders(t *testing.T) {
	m := sized(newModel(sample, "sample", "go"), 100, 50)
	if m.area.Value() != strings.TrimSuffix(sample, "") {
		t.Fatalf("value changed: %d tabs in, %d kept", strings.Count(sample, "\t"), strings.Count(m.area.Value(), "\t"))
	}
	fmt.Println(screen(m))
	for range 5 {
		m = press(m, "down")
	}
	for range 4 {
		m = press(m, "right")
		fmt.Println(m.stateLine())
	}
}

func TestPerCellMovement(t *testing.T) {
	m := sized(newModel("\t\tx\n\tyy", "t", "go"), 40, 10)
	m.handleKnob("f5")
	for range 6 {
		m = press(m, "right")
		fmt.Println(ansi.Strip(m.stateLine()))
	}
	m = press(m, "down")
	fmt.Println("after down:", ansi.Strip(m.stateLine()))
	m.handleKnob("f5")
}

func TestCursorOnTabRender(t *testing.T) {
	m := sized(newModel("\tx := 1", "t", "go"), 30, 4)
	for range cursorModes {
		first := strings.Split(m.area.View(), "\n")[0]
		reverse := strings.Index(first, "\x1b[7m")
		fmt.Printf("cursor %-5s %q  reverse video starts: %q\n", cursorModes[m.cursorMode], ansi.Strip(first)[:12], first[reverse:reverse+12])
		m.handleKnob("f4")
	}
}

var timingCases =[]struct {
	name    string
	content string
}{
	{"64 KiB Go", generatedGo(64 * 1024)},
	{"256 KiB Go", generatedGo(256 * 1024)},
	{"10,000 short lines", shortLines(10000)},
	{"60,000 short lines", shortLines(60000)},
}

func TestTiming(t *testing.T) {
	for _, c := range timingCases {
		for _, windowed := range []bool{false, true} {
			textarea.WindowedRender = windowed
			for _, mode := range []int{0, 1, 2} {
				m := newModel(c.content, c.name, "go")
				m.highlightMode = mode
				m.applyKnobs()
				m.relexIfChanged()
				m = sized(m, 120, 40)
				m.lexTime = 0
				var typing, views time.Duration
				for range 10 {
					begin := time.Now()
					m = press(m, "x")
					typing += time.Since(begin)
					begin = time.Now()
					_ = m.View()
					views += time.Since(begin)
				}
				fmt.Printf("%-19s windowed=%-5v %-9s lines %6d  keypress (Update+relex) %8s  View %8s  of which lex %8s\n",
					c.name, windowed, highlightModes[mode], m.area.LineCount(), ms(typing/10), ms(views/10), ms(m.lexTime))
			}
		}
	}
	textarea.WindowedRender = false
}

func TestStockTiming(t *testing.T) {
	for _, c := range timingCases {
		area := stock.New()
		area.Prompt = ""
		area.MaxHeight = 0
		area.SetWidth(120)
		area.SetHeight(35)
		area.SetValue(c.content)
		area.MoveToBegin()
		area.Focus()
		var typing, views time.Duration
		for range 10 {
			begin := time.Now()
			area, _ = area.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
			typing += time.Since(begin)
			begin = time.Now()
			_ = area.View()
			views += time.Since(begin)
		}
		fmt.Printf("%-19s stock bubbles v2.2.1, no highlighting, lines %6d  keypress %8s  View %8s\n",
			c.name, area.LineCount(), ms(typing/10), ms(views/10))
	}
}
