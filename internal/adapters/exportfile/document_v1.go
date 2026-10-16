package exportfile

type documentV1 struct {
	Format   string            `json:"format"`
	Version  int               `json:"version"`
	Folders  []folderRecordV1  `json:"folders"`
	Tags     []tagRecord       `json:"tags"`
	Snippets []snippetRecordV1 `json:"snippets"`
}

func (d documentV1) upgraded() document {
	return document{
		Format:   d.Format,
		Version:  formatVersion,
		Folders:  recordsOf(d.Folders, folderRecordV1.upgraded),
		Tags:     d.Tags,
		Snippets: recordsOf(d.Snippets, snippetRecordV1.upgraded),
	}
}
