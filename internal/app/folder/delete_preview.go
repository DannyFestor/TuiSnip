package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type DeletePreview struct {
	Folder         domain.Folder
	SubfolderCount int
	SnippetCount   int
}
