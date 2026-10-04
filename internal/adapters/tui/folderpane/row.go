package folderpane

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	rootLabel      = "◆ Root"
	expandedMarker = "▾ "
	leafMarker     = "  "
	indentPerLevel = "  "
)

type row struct {
	folderID     domain.FolderID
	label        string
	snippetCount int
}

func rowsOf(tree browse.Tree) []row {
	rows := []row{{folderID: domain.FolderID{}, label: rootLabel, snippetCount: tree.RootSnippetCount}}

	return appendNodes(rows, tree.Folders, 0)
}

func appendNodes(rows []row, nodes []browse.FolderNode, depth int) []row {
	for index := range nodes {
		node := &nodes[index]
		rows = append(rows, row{
			folderID:     node.Folder.ID(),
			label:        strings.Repeat(indentPerLevel, depth) + markerOf(node) + node.Folder.Name().String(),
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
