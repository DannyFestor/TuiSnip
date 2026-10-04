package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CreateInput struct {
	Name     string
	ParentID domain.FolderID
}
