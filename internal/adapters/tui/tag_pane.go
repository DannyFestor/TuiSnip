package tui

import "github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"

const (
	tagPaneTitle = "2 Tags"
	noTagsText   = "No Tags yet."
)

type tagPane struct{}

func (tagPane) body(styles look.Styles) string {
	return styles.Dim.Render(noTagsText)
}
