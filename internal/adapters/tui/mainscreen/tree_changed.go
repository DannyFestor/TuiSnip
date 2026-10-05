package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TreeChanged struct {
	Tree      browse.Tree
	Selecting domain.FolderID
}
