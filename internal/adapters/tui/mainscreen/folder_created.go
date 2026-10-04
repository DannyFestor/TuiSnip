package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderCreated struct {
	Tree browse.Tree
	ID   domain.FolderID
}
