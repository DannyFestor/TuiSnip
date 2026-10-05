package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type IDGenerator interface {
	NewFolderID() domain.FolderID
}
