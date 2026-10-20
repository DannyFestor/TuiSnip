package folderoutline

import "github.com/DannyFestor/TuiSnip/internal/app/trash"

const (
	subfolderMarker = "▾ "
	indentPerLevel  = "  "
)

func outlineLines(outline trash.Outline, indent string) []string {
	lines := make([]string, 0, len(outline.SnippetTitles))
	for _, title := range outline.SnippetTitles {
		lines = append(lines, indent+title)
	}

	for _, subfolder := range outline.Subfolders {
		lines = append(lines, indent+subfolderMarker+subfolder.Name)
		lines = append(lines, outlineLines(subfolder, indent+indentPerLevel)...)
	}

	return lines
}
