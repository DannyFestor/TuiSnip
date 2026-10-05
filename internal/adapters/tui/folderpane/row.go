package folderpane

import (
	"strconv"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	rootPrefix     = "◆ "
	rootName       = "Root"
	expandedMarker = "▾ "
	leafMarker     = "  "
	indentPerLevel = "  "
)

type row struct {
	folderID     domain.FolderID
	prefix       string
	name         string
	childIndent  string
	snippetCount int
}

type line struct {
	text string
	meta string
}

func rowsOf(tree browse.Tree) []row {
	rows := []row{{
		folderID:     domain.FolderID{},
		prefix:       rootPrefix,
		name:         rootName,
		childIndent:  "",
		snippetCount: tree.RootSnippetCount,
	}}

	return appendNodes(rows, tree.Folders, 0)
}

func appendNodes(rows []row, nodes []browse.FolderNode, depth int) []row {
	indent := strings.Repeat(indentPerLevel, depth)

	for index := range nodes {
		node := &nodes[index]
		rows = append(rows, row{
			folderID:     node.Folder.ID(),
			prefix:       indent + markerOf(node),
			name:         node.Folder.Name().String(),
			childIndent:  indent + indentPerLevel,
			snippetCount: node.SnippetCount,
		})
		rows = appendNodes(rows, node.Children, depth+1)
	}

	return rows
}

func markerOf(node *browse.FolderNode) string {
	if len(node.Children) == 0 {
		return leafMarker
	}

	return expandedMarker
}

func (r row) count() string {
	return strconv.Itoa(r.snippetCount)
}

func linesOf(rows []row) []line {
	lines := make([]line, 0, len(rows))
	for _, shown := range rows {
		lines = append(lines, line{text: shown.prefix + shown.name, meta: shown.count()})
	}

	return lines
}
