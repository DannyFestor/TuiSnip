package fragmentstrip

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	labelPadding = " "
	separator    = "│"
	ellipsis     = "…"
)

type Strip struct {
	styles  look.Styles
	labels  []string
	visible bool
	shown   int
}

func New(styles look.Styles) Strip {
	return Strip{styles: styles, labels: nil, visible: false, shown: 0}
}

func (s Strip) Of(fragments []domain.Fragment) Strip {
	tabs := make([]Tab, 0, len(fragments))
	for _, fragment := range fragments {
		tabs = append(tabs, tabOf(fragment))
	}

	return s.OfTabs(tabs)
}

func (s Strip) OfTabs(tabs []Tab) Strip {
	s.labels = labelsOf(tabs)
	s.visible = len(tabs) > 1 || (len(tabs) == 1 && tabs[0].Title.IsSet())
	s.shown = max(0, min(s.shown, len(s.labels)-1))

	return s
}

func (s Strip) Restyled(styles look.Styles) Strip {
	s.styles = styles

	return s
}

func (s Strip) Visible() bool {
	return s.visible
}

func (s Strip) Shown() int {
	return s.shown
}

func (s Strip) First() Strip {
	s.shown = 0

	return s
}

func (s Strip) Next() Strip {
	return s.Showing(s.shown + 1)
}

func (s Strip) Previous() Strip {
	return s.Showing(s.shown - 1)
}

func (s Strip) Showing(index int) Strip {
	if len(s.labels) == 0 {
		return s
	}

	s.shown = (index + len(s.labels)) % len(s.labels)

	return s
}

func (s Strip) Clicked(x int) Strip {
	start := 0

	for index, label := range s.labels {
		end := start + ansi.StringWidth(padded(label))
		if x >= start && x < end {
			return s.Showing(index)
		}

		start = end + ansi.StringWidth(separator)
	}

	return s
}

func (s Strip) View(width int) string {
	rendered := make([]string, 0, len(s.labels))
	for index, label := range s.labels {
		rendered = append(rendered, s.styleOf(index).Render(padded(label)))
	}

	return ansi.Truncate(strings.Join(rendered, s.styles.Dim.Render(separator)), width, ellipsis)
}

func (s Strip) styleOf(index int) lipgloss.Style {
	if index == s.shown {
		return s.styles.Focused.Cursor
	}

	return s.styles.Dim
}

func labelsOf(tabs []Tab) []string {
	labels := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		labels = append(labels, tab.Label())
	}

	return labels
}

func padded(label string) string {
	return labelPadding + label + labelPadding
}
