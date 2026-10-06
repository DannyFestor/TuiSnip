package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type MoveInput struct {
	FolderID domain.FolderID
	ParentID domain.FolderID
}
