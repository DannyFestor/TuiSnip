package browse

import "github.com/DannyFestor/TuiSnip/internal/domain"

type FolderNode struct {
	Folder       domain.Folder
	SnippetCount int
	Children     []FolderNode
}
