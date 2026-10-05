package folder

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SetDefaultLanguageInput struct {
	FolderID domain.FolderID
	Language string
}
