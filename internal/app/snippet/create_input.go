package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CreateInput struct {
	Title       string
	Description string
	Content     string
	FolderID    domain.FolderID
	Tags        []domain.Tag
}
