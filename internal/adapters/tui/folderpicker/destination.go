package folderpicker

import (
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/picker"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type destination struct {
	folderID domain.FolderID
	refused  bool
}

func destinationsIn(tree browse.Tree, moving domain.FolderID) []destination {
	root := destination{folderID: domain.FolderID{}, refused: false}

	return appendDestinations([]destination{root}, tree.Folders, false, moving)
}

func appendDestinations(
	destinations []destination, nodes []browse.FolderNode, insideMoving bool, moving domain.FolderID,
) []destination {
	for index := range nodes {
		node := &nodes[index]
		refused := insideMoving || node.Folder.ID() == moving
		destinations = append(destinations, destination{folderID: node.Folder.ID(), refused: refused})
		destinations = appendDestinations(destinations, node.Children, refused, moving)
	}

	return destinations
}

func choicesOf(destinations []destination, paths folderpath.Paths) []picker.Choice {
	choices := make([]picker.Choice, 0, len(destinations))
	for _, listed := range destinations {
		choices = append(choices, picker.Choice{Text: paths.Short(listed.folderID), Meta: ""})
	}

	return choices
}

func refusedIndexes(destinations []destination) []int {
	var refused []int

	for index, listed := range destinations {
		if listed.refused {
			refused = append(refused, index)
		}
	}

	return refused
}

func indexOf(destinations []destination, folderID domain.FolderID) int {
	return slices.IndexFunc(destinations, func(listed destination) bool { return listed.folderID == folderID })
}
