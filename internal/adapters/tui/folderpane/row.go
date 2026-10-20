package folderpane

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	rootPrefix      = "◆ "
	rootName        = "Root"
	trashPrefix     = "▪ "
	trashName       = "Trash"
	expandedMarker  = "▾ "
	collapsedMarker = "▸ "
	leafMarker      = "  "
	indentPerLevel  = "  "
)

type row struct {
	folderID     domain.FolderID
	folder       domain.Folder
	prefix       string
	name         string
	childIndent  string
	snippetCount int
	collapsible  bool
	trash        bool
}

type line struct {
	text string
	meta string
}

func rowsOf(tree browse.Tree, collapsed collapsedSet) []row {
	rows := []row{{
		folderID:     domain.FolderID{},
		folder:       domain.Folder{},
		prefix:       rootPrefix,
		name:         rootName,
		childIndent:  "",
		snippetCount: tree.RootSnippetCount,
		collapsible:  false,
		trash:        false,
	}}

	return append(appendNodes(rows, tree.Folders, 0, collapsed), trashRow(tree.TrashCount))
}

func trashRow(count int) row {
	return row{
		folderID:     domain.FolderID{},
		folder:       domain.Folder{},
		prefix:       trashPrefix,
		name:         trashName,
		childIndent:  "",
		snippetCount: count,
		collapsible:  false,
		trash:        true,
	}
}

func appendNodes(rows []row, nodes []browse.FolderNode, depth int, collapsed collapsedSet) []row {
	indent := strings.Repeat(indentPerLevel, depth)

	for index := range nodes {
		node := &nodes[index]
		isCollapsed := collapsed.has(node.Folder.ID())
		rows = append(rows, row{
			folderID:     node.Folder.ID(),
			folder:       node.Folder,
			prefix:       indent + markerOf(node, isCollapsed),
			name:         node.Folder.Name().String(),
			childIndent:  indent + indentPerLevel,
			snippetCount: node.SnippetCount,
			collapsible:  len(node.Children) > 0,
			trash:        false,
		})

		if !isCollapsed {
			rows = appendNodes(rows, node.Children, depth+1, collapsed)
		}
	}

	return rows
}

func markerOf(node *browse.FolderNode, isCollapsed bool) string {
	switch {
	case len(node.Children) == 0:
		return leafMarker
	case isCollapsed:
		return collapsedMarker
	}

	return expandedMarker
}

func (r row) markerHolds(x int) bool {
	markerEnd := ansi.StringWidth(r.prefix)

	return r.collapsible && x >= markerEnd-ansi.StringWidth(expandedMarker) && x < markerEnd
}

func (r row) count() string {
	return strconv.Itoa(r.snippetCount)
}

func linesOf(rows []row) []line {
	lines := make([]line, 0, len(rows))
	for index := range rows {
		shown := &rows[index]
		lines = append(lines, line{text: shown.prefix + shown.name, meta: shown.count()})
	}

	return lines
}
