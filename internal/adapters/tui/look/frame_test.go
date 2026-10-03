package look_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
)

func TestFrame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		body  string
		outer look.Size
		want  []string
	}{
		{
			name:  "pads the body to the inner box",
			title: "Tags",
			body:  "go",
			outer: look.Size{Width: 10, Height: 4},
			want:  []string{"╭ Tags ──╮", "│go      │", "│        │", "╰────────╯"},
		},
		{
			name:  "cuts body lines below the box and truncates wide ones",
			title: "Tags",
			body:  "first line\nsecond\nthird",
			outer: look.Size{Width: 8, Height: 4},
			want:  []string{"╭ Tags ╮", "│first…│", "│second│", "╰──────╯"},
		},
		{
			name:  "truncates a title wider than the border",
			title: "Snippets",
			body:  "",
			outer: look.Size{Width: 8, Height: 2},
			want:  []string{"╭ Sni… ╮", "╰──────╯"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := look.Frame(plainFrame(), tt.title, tt.body, tt.outer)

			assert.Equal(t, strings.Join(tt.want, "\n"), got)
		})
	}
}

func TestFitWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		line  string
		width int
		want  string
	}{
		{name: "pads a short line", line: "go", width: 5, want: "go   "},
		{name: "keeps a line that fits exactly", line: "go mod", width: 6, want: "go mod"},
		{name: "truncates a long line with an ellipsis", line: "go mod tidy", width: 6, want: "go mo…"},
		{name: "counts wide characters by cell", line: "日本", width: 5, want: "日本 "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, look.FitWidth(tt.line, tt.width))
		})
	}
}

func TestRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		meta  string
		width int
		want  string
	}{
		{name: "puts the meta at the right edge", text: "curl", meta: "Bash", width: 12, want: "curl    Bash"},
		{
			name:  "keeps one space when the text fills the rest",
			text:  "shutdown",
			meta:  "Go",
			width: 11,
			want:  "shutdown Go",
		},
		{
			name:  "truncates the text to keep the meta",
			text:  "graceful shutdown",
			meta:  "Go",
			width: 11,
			want:  "gracefu… Go",
		},
		{
			name:  "drops the text when the meta alone is too wide",
			text:  "x",
			meta:  "JavaScript",
			width: 6,
			want:  " Java…",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, look.Row(tt.text, tt.meta, tt.width))
		})
	}
}

func plainFrame() look.FrameStyle {
	return look.FrameStyle{Border: lipgloss.NewStyle(), Title: lipgloss.NewStyle(), Cursor: lipgloss.NewStyle()}
}
