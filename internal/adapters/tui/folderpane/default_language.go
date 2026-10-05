package folderpane

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func (p Pane) DefaultLanguageOf(id domain.FolderID) value.Language {
	folder, found := folderIn(p.tree.Folders, id)
	if !found {
		return value.PlainText()
	}

	return folder.DefaultLanguage()
}

func folderIn(nodes []browse.FolderNode, id domain.FolderID) (domain.Folder, bool) {
	for index := range nodes {
		node := &nodes[index]
		if node.Folder.ID() == id {
			return node.Folder, true
		}

		folder, found := folderIn(node.Children, id)
		if found {
			return folder, true
		}
	}

	return domain.Folder{}, false
}
