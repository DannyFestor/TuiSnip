package folderpane

import (
	"maps"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type collapsedSet map[domain.FolderID]struct{}

func collapsedSetOf(ids []domain.FolderID) collapsedSet {
	set := make(collapsedSet, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}

	return set
}

func (s collapsedSet) has(id domain.FolderID) bool {
	_, ok := s[id]

	return ok
}

func (s collapsedSet) toggled(id domain.FolderID) collapsedSet {
	if s.has(id) {
		return s.without(id)
	}

	next := maps.Clone(s)
	next[id] = struct{}{}

	return next
}

func (s collapsedSet) without(ids ...domain.FolderID) collapsedSet {
	next := maps.Clone(s)
	for _, id := range ids {
		delete(next, id)
	}

	return next
}

func (s collapsedSet) hiding(tree browse.Tree, id domain.FolderID) []domain.FolderID {
	ancestors, _ := ancestorsOf(tree.Folders, id)

	return slices.DeleteFunc(ancestors, func(ancestor domain.FolderID) bool { return !s.has(ancestor) })
}

func ancestorsOf(nodes []browse.FolderNode, id domain.FolderID) ([]domain.FolderID, bool) {
	for index := range nodes {
		node := &nodes[index]
		if node.Folder.ID() == id {
			return nil, true
		}

		ancestors, found := ancestorsOf(node.Children, id)
		if found {
			return append(ancestors, node.Folder.ID()), true
		}
	}

	return nil, false
}

func (s collapsedSet) within(tree browse.Tree) []domain.FolderID {
	return s.appendWithin(make([]domain.FolderID, 0, len(s)), tree.Folders)
}

func (s collapsedSet) appendWithin(ids []domain.FolderID, nodes []browse.FolderNode) []domain.FolderID {
	for index := range nodes {
		node := &nodes[index]
		if s.has(node.Folder.ID()) {
			ids = append(ids, node.Folder.ID())
		}

		ids = s.appendWithin(ids, node.Children)
	}

	return ids
}
