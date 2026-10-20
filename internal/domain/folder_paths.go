package domain

import "slices"

type FolderPaths struct {
	folders map[FolderID]Folder
	marks   map[FolderID]TrashMark
}

func NewFolderPaths(folders []Folder, trash Trash) FolderPaths {
	byID := make(map[FolderID]Folder, len(folders))
	for _, folder := range folders {
		byID[folder.ID()] = folder
	}

	return FolderPaths{folders: byID, marks: trash.Folders}
}

func (p FolderPaths) To(folderID FolderID) []string {
	lineage := p.lineageOf(folderID)
	if len(lineage) == 0 {
		return []string{}
	}

	names := p.marks[lineage[len(lineage)-1].ID()].FormerPath().Strings()
	for _, folder := range slices.Backward(lineage) {
		names = append(names, folder.Name().String())
	}

	return names
}

func (p FolderPaths) lineageOf(folderID FolderID) []Folder {
	var lineage []Folder

	visited := make(map[FolderID]bool)

	for folder, known := p.folders[folderID]; known && !visited[folderID]; folder, known = p.folders[folderID] {
		visited[folderID] = true

		lineage = append(lineage, folder)
		folderID = folder.ParentID()
	}

	return lineage
}
