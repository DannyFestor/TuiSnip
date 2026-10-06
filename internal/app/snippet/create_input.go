package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CreateInput struct {
	Title       string
	Description string
	Language    string
	Content     string
	FolderID    domain.FolderID
	Tags        []domain.Tag
	NewTags     []string
}

func (in CreateInput) raw() rawFields {
	return rawFields{
		title:       in.Title,
		description: in.Description,
		language:    in.Language,
		content:     in.Content,
		newTags:     in.NewTags,
	}
}
