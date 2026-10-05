package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Destination struct {
	Selection browseselection.Selection
	Tags      []domain.Tag
}

func (d Destination) folderID() domain.FolderID {
	folderID, _ := d.Selection.Folder()

	return folderID
}
