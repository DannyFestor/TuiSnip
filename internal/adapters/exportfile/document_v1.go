package exportfile

type documentV1 struct {
	Format   string          `json:"format"`
	Version  int             `json:"version"`
	Folders  []folderFields  `json:"folders"`
	Tags     []tagRecord     `json:"tags"`
	Snippets []snippetFields `json:"snippets"`
}

func (d documentV1) upgraded() document {
	return document{
		Format:   d.Format,
		Version:  formatVersion,
		Folders:  recordsOf(d.Folders, upgradedFolder),
		Tags:     d.Tags,
		Snippets: recordsOf(d.Snippets, upgradedSnippet),
	}
}

func upgradedFolder(fields folderFields) folderRecord {
	return folderRecord{folderFields: fields, DeletedAt: nil, FormerParentPath: []string{}}
}

func upgradedSnippet(fields snippetFields) snippetRecord {
	return snippetRecord{snippetFields: fields, DeletedAt: nil, FormerFolderPath: []string{}}
}
