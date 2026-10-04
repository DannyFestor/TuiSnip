package folderpath

import (
	"slices"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	Root      = "Root"
	separator = " / "
)

type Paths struct {
	names map[domain.FolderID][]string
}

func New(tree browse.Tree) Paths {
	names := make(map[domain.FolderID][]string)
	collectNames(names, nil, tree.Folders)

	return Paths{names: names}
}

func (p Paths) Folder(id domain.FolderID) string {
	return strings.Join(append([]string{Root}, p.namesOf(id)...), separator)
}

func (p Paths) Of(snippet domain.Snippet) string {
	return p.Folder(snippet.FolderID())
}

func (p Paths) Compact(snippet domain.Snippet) string {
	if snippet.AtRoot() {
		return Root
	}

	return strings.Join(p.namesOf(snippet.FolderID()), separator)
}

func (p Paths) namesOf(id domain.FolderID) []string {
	if id.IsNil() {
		return nil
	}

	names, ok := p.names[id]
	if !ok {
		return []string{look.Ellipsis}
	}

	return names
}

func collectNames(names map[domain.FolderID][]string, parent []string, nodes []browse.FolderNode) {
	for index := range nodes {
		node := &nodes[index]
		path := append(slices.Clone(parent), node.Folder.Name().String())
		names[node.Folder.ID()] = path
		collectNames(names, path, node.Children)
	}
}
