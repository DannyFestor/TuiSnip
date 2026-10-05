package tag

import "github.com/DannyFestor/TuiSnip/internal/domain"

type RenameInput struct {
	TagID domain.TagID
	Name  string
}
