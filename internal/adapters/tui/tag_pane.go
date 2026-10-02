package tui

const (
	tagPaneTitle = "2 Tags"
	noTagsText   = "No Tags yet."
)

type tagPane struct{}

func (tagPane) body(styles styleSet) string {
	return styles.dim.Render(noTagsText)
}
