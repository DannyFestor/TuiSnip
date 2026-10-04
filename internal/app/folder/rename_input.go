package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type RenameInput struct {
	FolderID domain.FolderID
	Name     string
}
